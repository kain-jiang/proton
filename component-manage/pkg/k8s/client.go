package k8s

import (
	"context"
	"fmt"
	"time"

	errors2 "errors"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

var (
	GVRMariaDB = schema.GroupVersionResource{
		Group:    "rds.proton.aishu.cn",
		Version:  "v1",
		Resource: "rdsmariadbclusters",
	}
	GVRMongoDB = schema.GroupVersionResource{
		Group:    "mongodb.proton.aishu.cn",
		Version:  "v1",
		Resource: "mongodboperators",
	}
	GVRNebula = schema.GroupVersionResource{
		Group:    "apps.nebula-graph.io",
		Version:  "v1alpha1",
		Resource: "nebulaclusters",
	}
)

type (
	cli struct {
		clientSet  kubernetes.Interface
		dynamicSet dynamic.Interface
	}

	Client interface {
		GetMasterNodes() ([]v1.Node, error)
		ListNodes() ([]v1.Node, error)
		SelfNameSpace() string
		// ConfigMapGet 忽略不存在的错误
		ConfigMapGet(name, ns string) (map[string]string, error)
		ConfigMapDel(name, ns string) error
		ConfigMapSet(name, ns string, data map[string]string) error

		SecretGet(name, ns string) (map[string][]byte, error)
		SecretSet(name, ns string, data map[string][]byte) error
		SecretDel(name, ns string) error
		SecretStringSet(name, ns string, data map[string]string) error
		SecretExist(name, ns string) (bool, error)
		CheckJobStatus(namespace, JobName string) (bool, error)
		DeleteJob(namespace, JobName string) error

		CustomResourceGet(gvr schema.GroupVersionResource, name, ns string) (map[string]any, error)
		CustomResourceUpdate(gvr schema.GroupVersionResource, name, ns string, obj map[string]any) (map[string]any, error)
		CustomResourceCreate(gvr schema.GroupVersionResource, name, ns string, obj map[string]any) (map[string]any, error)
		CustomResourceSet(gvr schema.GroupVersionResource, name, ns string, obj map[string]any) (map[string]any, error)
		CustomResourceDelete(gvr schema.GroupVersionResource, name, ns string) error

		WaitForDeploymentReady(timeout time.Duration, name, ns string) error
		WaitForStatefulSetReady(timeout time.Duration, name, ns string) error
	}
)

func (c *cli) GetMasterNodes() ([]v1.Node, error) {
	nodeList, err := c.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/master="})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

func (c *cli) ListNodes() ([]v1.Node, error) {
	nodeList, err := c.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

func (c *cli) ConfigMapGet(name, ns string) (map[string]string, error) {
	d, err := c.clientSet.CoreV1().ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return d.Data, nil
}

func (c *cli) ConfigMapDel(name, ns string) error {
	err := c.clientSet.CoreV1().ConfigMaps(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		err = nil
	}
	return err
}

func (c *cli) SecretDel(name, ns string) error {
	err := c.clientSet.CoreV1().Secrets(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		err = nil
	}
	return err
}

func (c *cli) ConfigMapSet(name, ns string, data map[string]string) error {
	cm, err := c.clientSet.CoreV1().ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		cm.Data = data
		_, err := c.clientSet.CoreV1().ConfigMaps(ns).Update(context.TODO(), cm, metav1.UpdateOptions{})
		return err
	} else if errors.IsNotFound(err) {
		// Create
		cm = &v1.ConfigMap{Data: data}
		cm.SetName(name)
		cm.SetNamespace(ns)
		_, err := c.clientSet.CoreV1().ConfigMaps(ns).Create(context.TODO(), cm, metav1.CreateOptions{})
		return err
	} else {
		return err
	}
}

func (c *cli) SelfNameSpace() string {
	return SelfNameSpace()
}

func (c *cli) SecretExist(name, ns string) (bool, error) {
	_, err := c.clientSet.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// 出现不存在的错误
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *cli) SecretGet(name, ns string) (map[string][]byte, error) {
	d, err := c.clientSet.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return map[string][]byte{}, nil
		} else {
			return nil, err
		}
	}
	if d.Data == nil {
		return map[string][]byte{}, nil
	}
	return d.Data, nil
}

func (c *cli) SecretStringSet(name, ns string, data map[string]string) error {
	secet, err := c.clientSet.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		secet.StringData = data
		_, err := c.clientSet.CoreV1().Secrets(ns).Update(context.TODO(), secet, metav1.UpdateOptions{})
		return err
	} else if errors.IsNotFound(err) {
		// Create
		secet = &v1.Secret{StringData: data}
		secet.SetName(name)
		secet.SetNamespace(ns)
		_, err := c.clientSet.CoreV1().Secrets(ns).Create(context.TODO(), secet, metav1.CreateOptions{})
		return err
	} else {
		return err
	}
}

func (c *cli) SecretSet(name, ns string, data map[string][]byte) error {
	secet, err := c.clientSet.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		secet.Data = data
		_, err := c.clientSet.CoreV1().Secrets(ns).Update(context.TODO(), secet, metav1.UpdateOptions{})
		return err
	} else if errors.IsNotFound(err) {
		// Create
		secet = &v1.Secret{Data: data}
		secet.SetName(name)
		secet.SetNamespace(ns)
		_, err := c.clientSet.CoreV1().Secrets(ns).Create(context.TODO(), secet, metav1.CreateOptions{})
		return err
	} else {
		return err
	}
}

func (c *cli) CheckJobStatus(namespace, jobName string) (bool, error) {
	// 获取指定job是否完成
	job, err := c.clientSet.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		//if errors.IsNotFound(err) {
		//	// 出现不存在的错误
		//	return true, nil
		//} else {
		return false, err
		//}
	}

	failedTaskCount := 0
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == v1.ConditionTrue {
			failedTaskCount++
		}
	}

	// 检查失败任务数量
	if failedTaskCount >= 1 {
		// errMsg := fmt.Sprintf("Job %s has %d failed tasks", jobName, failedTaskCount)
		errMsg := fmt.Errorf("Job %s has %d failed tasks", jobName, failedTaskCount)
		return false, errMsg
	}

	// 检查 Job 是否已经完成
	if job.Status.CompletionTime != nil {
		return true, nil
	} else {
		return false, nil
	}
}

func (c *cli) DeleteJob(namespace, JobName string) error {
	err := c.clientSet.BatchV1().Jobs(namespace).Delete(context.TODO(), JobName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	podClient := c.clientSet.CoreV1().Pods(namespace)
	podList, err := podClient.List(context.TODO(), metav1.ListOptions{LabelSelector: fmt.Sprintf("job-name= %s", JobName)})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	for _, pod := range podList.Items {
		err := podClient.Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	return nil
}

func (c *cli) CustomResourceGet(gvr schema.GroupVersionResource, name, ns string) (obj map[string]any, err error) {
	cr, err := c.dynamicSet.Resource(gvr).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return cr.UnstructuredContent(), err
}

func (c *cli) CustomResourceCreate(gvr schema.GroupVersionResource, name, ns string, obj map[string]any) (map[string]any, error) {
	cr, err := c.dynamicSet.Resource(gvr).Namespace(ns).Create(context.TODO(), &unstructured.Unstructured{Object: obj}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return cr.UnstructuredContent(), err
}

func (c *cli) CustomResourceUpdate(gvr schema.GroupVersionResource, name, ns string, obj map[string]any) (map[string]any, error) {
	cr, err := c.dynamicSet.Resource(gvr).Namespace(ns).Update(context.TODO(), &unstructured.Unstructured{Object: obj}, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return cr.UnstructuredContent(), err
}

func (c *cli) CustomResourceSet(gvr schema.GroupVersionResource, name, ns string, obj map[string]any) (map[string]any, error) {
	cr, err := c.CustomResourceGet(gvr, name, ns)
	if err != nil {
		return nil, err
	}
	if cr == nil {
		return c.CustomResourceCreate(gvr, name, ns, obj)
	}
	// 更新资源需要 metadata.resourceVersion
	obj["metadata"].(map[string]any)["resourceVersion"] = cr["metadata"].(map[string]any)["resourceVersion"]
	return c.CustomResourceUpdate(gvr, name, ns, obj)
}

func (c *cli) CustomResourceDelete(gvr schema.GroupVersionResource, name, ns string) error {
	err := c.dynamicSet.Resource(gvr).Namespace(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}

func (c *cli) WaitForDeploymentReady(timeout time.Duration, name, ns string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return retry.OnError(retry.DefaultRetry, func(error) bool { return true }, func() error {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("timed out waiting for deployment %s to be ready", name)
			case <-ticker.C:
				deployment, err := c.clientSet.AppsV1().Deployments(ns).Get(context.TODO(), name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
					return nil
				}
			}
		}
	})
}

func (c *cli) WaitForStatefulSetReady(timeout time.Duration, name, ns string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return retry.OnError(retry.DefaultRetry, func(error) bool { return true }, func() error {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("timed out waiting for deployment %s to be ready", name)
			case <-ticker.C:
				statefulSet, err := c.clientSet.AppsV1().StatefulSets(ns).Get(context.TODO(), name, metav1.GetOptions{})
				if err != nil {
					return err
				}
				if statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas {
					return nil
				}
			}
		}
	})
}

func New() Client {
	config, err := rest.InClusterConfig()

	if errors2.Is(err, rest.ErrNotInCluster) {
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	}
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	dynamicSet, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return &cli{
		clientSet:  clientSet,
		dynamicSet: dynamicSet,
	}
}
