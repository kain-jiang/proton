package protonk8s

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
)

// Config represents the configuration for the templates
type Config struct {
	Version          string
	CurrentVersion   string
	TemplateYAML     string
	CalicoVethMTU    string
	ImageRepository  string
	IPv6Interface    string
	PodNetworkCIDRv6 string
	PodNetworkCIDRv4 string
}

var (
	nameSpace        = "kube-system"
	configMapName    = "calico-config"
	DSName           = "calico-node"
	DefaultImageRepo = "registry.aishu.cn:15000/public"
	DefaultMTU       = "1440"
	SupportedVersion = map[string]string{"v3.25.2": calicoV3252YamlTemplate}
	log              = logger.NewLogger()
)

func GetSupportedVersion() []string {
	var versions []string
	for k := range SupportedVersion {
		versions = append(versions, k)
	}
	return versions
}
func pushImages(k kubernetes.Interface) error {
	cfg, err := configuration.LoadFromKubernetes(context.Background(), k)
	if err != nil {
		log.Errorf("unable load cluster conf: %v", err)
	}
	c := &cr.Cr{
		Logger:        logger.NewLogger(),
		ClusterConf:   cfg,
		PrePullImages: false,
	}
	if err := c.Apply(); err != nil {
		return err
	}

	return nil
}

func (c *Config) GetCurrentCalicoConfig(k kubernetes.Interface) error {
	cm, err := k.CoreV1().ConfigMaps(nameSpace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap error: %w", err)
	}
	calicoVersion, exists := cm.ObjectMeta.Annotations["calicoVersion"]
	if exists {
		c.CurrentVersion = calicoVersion
	}
	return nil
}

func (c *Config) getImageRepository(k kubernetes.Interface) error {
	ctx := context.TODO()
	clusterCfg, err := configuration.LoadFromKubernetes(ctx, k)
	if err != nil {
		return fmt.Errorf("cannot get proton cluster config from kubernetes, error: %w", err)
	}
	if clusterCfg.Cr.External != nil {
		c.ImageRepository = clusterCfg.Cr.External.ImageRepository()
	} else {
		c.ImageRepository = DefaultImageRepo
	}
	return nil
}

func (c *Config) getPodNetworkInfo(k kubernetes.Interface) error {
	cm, err := k.CoreV1().ConfigMaps(nameSpace).Get(context.TODO(), configMapName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get configmap MTU error: %w", err)
	}
	vethMtuValue, ok := cm.Data["veth_mtu"]
	if !ok {
		c.CalicoVethMTU = vethMtuValue
	} else {
		c.CalicoVethMTU = DefaultMTU
	}

	calicoVersion, exists := cm.ObjectMeta.Annotations["calicoVersion"]
	if exists {
		c.Version = calicoVersion
	}

	daemonSet, err := k.AppsV1().DaemonSets(nameSpace).Get(context.TODO(), DSName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get calico daemonset, error: %w", err)
	}

	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		for _, envVar := range container.Env {
			if envVar.Name == "CALICO_IPV6POOL_CIDR" {
				c.PodNetworkCIDRv6 = envVar.Value
			} else if envVar.Name == "CALICO_IPV4POOL_CIDR" {
				c.PodNetworkCIDRv4 = envVar.Value
			} else if envVar.Name == "IP6_AUTODETECTION_METHOD" {
				c.IPv6Interface = envVar.Value
			}
		}
	}
	return nil
}

func (c *Config) renderTemplate(t string) error {
	tmpl, err := template.New("").Parse(t)
	if err != nil {
		return err
	}

	var renderTemplate bytes.Buffer
	if err := tmpl.Execute(&renderTemplate, c); err != nil {
		return err
	}
	c.TemplateYAML = renderTemplate.String()
	return nil
}

func decodeYAML(t string) ([]runtime.Object, error) {
	var objects []runtime.Object
	yamlParts := strings.Split(t, "---")
	for _, part := range yamlParts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		obj := &unstructured.Unstructured{}
		err := yaml.Unmarshal([]byte(part), obj)
		if err != nil {
			return nil, err
		}

		concreteObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil, err
		}

		var runtimeObj runtime.Object
		switch obj.GroupVersionKind().Kind {
		case "ServiceAccount":
			runtimeObj = &corev1.ServiceAccount{}
		case "PodDisruptionBudget":
			runtimeObj = &policyv1.PodDisruptionBudget{}
		case "ConfigMap":
			runtimeObj = &corev1.ConfigMap{}
		case "Deployment":
			runtimeObj = &appsv1.Deployment{}
		case "Service":
			runtimeObj = &corev1.Service{}
		case "Ingress":
			runtimeObj = &networkingv1.Ingress{}
		case "CronJob":
			runtimeObj = &batchv1beta1.CronJob{}
		case "Secret":
			runtimeObj = &corev1.Secret{}
		case "PersistentVolumeClaim":
			runtimeObj = &corev1.PersistentVolumeClaim{}
		case "CustomResourceDefinition":
			runtimeObj = &extv1.CustomResourceDefinition{}
		case "ClusterRole":
			runtimeObj = &rbacv1.ClusterRole{}
		case "ClusterRoleBinding":
			runtimeObj = &rbacv1.ClusterRoleBinding{}
		case "DaemonSet":
			runtimeObj = &appsv1.DaemonSet{}
		default:
			return nil, fmt.Errorf("unsupported kind: %s", obj.GroupVersionKind().Kind)
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(concreteObj, runtimeObj)
		if err != nil {
			return nil, fmt.Errorf("error converting unstructured to %s, %w", obj.GroupVersionKind().Kind, err)
		}

		objects = append(objects, runtimeObj)
	}

	return objects, nil
}

func (c *Config) applyYAML(extclient *apiextensionsv1.ApiextensionsV1Client, k kubernetes.Interface) error {
	log.Println("applying calico yaml")
	objects, err := decodeYAML(c.TemplateYAML)
	if err != nil {
		log.Println("decode yaml failed")
		return err
	}

	// 应用YAML文件
	for _, obj := range objects {
		switch o := obj.(type) {
		case *corev1.ServiceAccount:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico ServiceAccount %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.CoreV1().ServiceAccounts(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.CoreV1().ServiceAccounts(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.CoreV1().ServiceAccounts(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.CoreV1().ServiceAccounts(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *policyv1.PodDisruptionBudget:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico PodDisruptionBudget %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.PolicyV1().PodDisruptionBudgets(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.PolicyV1().PodDisruptionBudgets(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.PolicyV1().PodDisruptionBudgets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.PolicyV1().PodDisruptionBudgets(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *corev1.ConfigMap:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico ConfigMap %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.CoreV1().ConfigMaps(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.CoreV1().ConfigMaps(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.CoreV1().ConfigMaps(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.CoreV1().ConfigMaps(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *appsv1.Deployment:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico Deployment %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.AppsV1().Deployments(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.AppsV1().Deployments(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.AppsV1().Deployments(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.AppsV1().Deployments(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *corev1.Service:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico Services %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.CoreV1().Services(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.CoreV1().Services(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.CoreV1().Services(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.CoreV1().Services(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *networkingv1.Ingress:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico Ingress %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.NetworkingV1().Ingresses(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.NetworkingV1().Ingresses(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.NetworkingV1().Ingresses(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.NetworkingV1().Ingresses(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *batchv1beta1.CronJob:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico CronJob %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.BatchV1beta1().CronJobs(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.BatchV1beta1().CronJobs(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.BatchV1beta1().CronJobs(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.BatchV1beta1().CronJobs(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *corev1.Secret:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico Secret %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.CoreV1().Secrets(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.CoreV1().Secrets(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.CoreV1().Secrets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.CoreV1().Secrets(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *corev1.PersistentVolumeClaim:
			if o.Name == "" {
				continue
			}
			if o.Namespace == "" {
				o.Namespace = nameSpace
			}
			log.Printf("applying calico PersistentVolumeClaim %s/%s", o.Namespace, o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.CoreV1().PersistentVolumeClaims(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.CoreV1().PersistentVolumeClaims(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.CoreV1().PersistentVolumeClaims(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
			// _, err = k.CoreV1().PersistentVolumeClaims(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
		case *extv1.CustomResourceDefinition:
			if o.Name == "" {
				continue
			}
			log.Printf("applying calico CustomResourceDefinition %s", o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := extclient.CustomResourceDefinitions().Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = extclient.CustomResourceDefinitions().Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = extclient.CustomResourceDefinitions().Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
		case *rbacv1.ClusterRole:
			if o.Name == "" {
				continue
			}
			log.Printf("applying calico ClusterRole %s", o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.RbacV1().ClusterRoles().Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.RbacV1().ClusterRoles().Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.RbacV1().ClusterRoles().Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
		case *rbacv1.ClusterRoleBinding:
			if o.Name == "" {
				continue
			}
			log.Printf("applying calico ClusterRoleBinding %s", o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.RbacV1().ClusterRoleBindings().Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.RbacV1().ClusterRoleBindings().Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.RbacV1().ClusterRoleBindings().Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
		case *appsv1.DaemonSet:
			if o.Name == "" {
				continue
			}
			log.Printf("applying calico DaemonSet %s", o.Name)
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				existingObj, err := k.AppsV1().DaemonSets(o.Namespace).Get(context.TODO(), o.Name, metav1.GetOptions{})
				if errors.IsNotFound(err) {
					// Object does not exist, create it
					_, err = k.AppsV1().DaemonSets(o.Namespace).Create(context.TODO(), o, metav1.CreateOptions{})
					if err != nil {
						return err
					}
					return nil
				} else if err != nil {
					return err
				}

				// Object exists, update it
				if existingObj != nil {
					o.SetResourceVersion(existingObj.GetResourceVersion())
					_, err = k.AppsV1().DaemonSets(o.Namespace).Update(context.TODO(), o, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
				return nil
			})
			if retryErr != nil {
				return retryErr
			}
		default:
			return fmt.Errorf("unsupported object type: %T", obj)
		}
	}

	return nil
}

func (c *Config) changeIPIPMode() error {
	// 构建kubectl patch命令字符串
	cmdStr := `kubectl patch ippool/default-ipv4-ippool --type='json' -p="[{\"op\":\"replace\", \"path\":\"/spec/ipipMode\", \"value\":\"Always\"}]"`

	// 创建命令对象
	cmd := exec.CommandContext(context.Background(), "sh", "-c", cmdStr)

	// 执行命令
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing kubectl command: %v\nOutput: %s", err, string(out))
	}

	log.Printf("update ipipMode to Always success")

	return nil
}

func (c *Config) CalicoUpgrade(extclient *apiextensionsv1.ApiextensionsV1Client, k kubernetes.Interface) error {
	if err := c.GetCurrentCalicoConfig(k); err != nil {
		return fmt.Errorf("failed to get current calico version: %w", err)
	}
	if c.CurrentVersion >= c.Version {
		log.Printf("Calico is already up to date. Version: %s\n", c.Version)
		return nil
	}
	if err := c.getPodNetworkInfo(k); err != nil {
		return fmt.Errorf("failed to get pod network info: %w", err)
	}
	if err := c.getImageRepository(k); err != nil {
		return fmt.Errorf("failed to get image repository info: %w", err)
	}

	if err := c.renderTemplate(SupportedVersion[c.Version]); err != nil {
		return fmt.Errorf("failed to convert calico template yaml, error: %w", err)
	}

	if err := pushImages(k); err != nil {
		return fmt.Errorf("failed to push images, error: %w", err)
	}

	if err := c.applyYAML(extclient, k); err != nil {
		return fmt.Errorf("failed to apply calico template yaml, : %w", err)
	}

	if err := c.changeIPIPMode(); err != nil {
		return fmt.Errorf("failed to change ipip mode, error: %w", err)
	}

	log.Println("calico upgrade success.")
	return nil
}
