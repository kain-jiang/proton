package kafka

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

const (
	// kafka 的 Chart 名称
	ChartName = "proton-kafka"
	// kafka 的 Helm release 名称
	ReleaseName = "kafka"
	// kafka 的 Helm release 所在的命名空间
	ReleaseNamespace = "resource"
	KafkaServiceName = "kafka-headless"

	KafkaServicePort = 9097

	KafkaDefaultSSLUser     = "FAKE_USERNAME"
	KafkaDefaultSSLPassword = "FAKE_PASSWORD"
)

var log = logger.NewLogger()

type KafkaManager struct {
	// kafka Spec
	spec *configuration.Kafka

	// 节点访问配置，用于生成 SSH 客户端配置
	hosts []configuration.Node

	// registry 地址
	registry string

	// Helm3 Client
	helm3 helm3.Client

	// service-package 的路径
	servicePackage string
	// chart 列表
	charts servicepackage.Charts

	// oldConfig 旧配置
	oldConf *configuration.Kafka
}

// 创建 KafkaManager
func New(spec *configuration.Kafka) *KafkaManager {
	return &KafkaManager{
		spec: spec,
	}
}

func (k *KafkaManager) Helm3(helm3 helm3.Client) *KafkaManager {
	k.helm3 = helm3
	return k
}

// 设置节点信息，用于通过 ssh 远程创建数据目录
func (k *KafkaManager) Hosts(hosts []configuration.Node) *KafkaManager {
	k.hosts = hosts
	return k
}

// 设置 Registry 地址
func (k *KafkaManager) Registry(registry string) *KafkaManager {
	k.registry = registry
	return k
}

// 设置 service-package 的路径
func (k *KafkaManager) ServicePackage(servicePackage string) *KafkaManager {
	k.servicePackage = servicePackage
	return k
}

// 设置 chart 列表
func (k *KafkaManager) Charts(charts servicepackage.Charts) *KafkaManager {
	k.charts = charts
	return k
}

// 设置 oldConfig 旧配置
func (k *KafkaManager) OldConfig(oldConf *configuration.Kafka) *KafkaManager {
	k.oldConf = oldConf
	return k
}

func (k *KafkaManager) Apply() error {
	var ctx = context.TODO()
	// 创建数据目录
	for _, host := range k.spec.Hosts {

		f := ecms.NewForHost(host).Files()

		if info, err := f.Stat(ctx, k.spec.Data_path); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			log.Printf("host[%s] create directory %s", host, k.spec.Data_path)
			if err := f.Create(ctx, k.spec.Data_path, true, nil); err != nil {
				return err
			}
		} else if !info.IsDir() {
			return fmt.Errorf("host[%s] %s is not a directory", host, k.spec.Data_path)
		}
	}
	// 向 helm client 注册安装命令
	return k.apply()
}

func (k *KafkaManager) apply() error {
	log.Infof("Applying release=%s chart=%s ", ReleaseName, ChartName)

	if err := k.UpgradeOrInstall(); err != nil {
		return fmt.Errorf("unable to upgrade release %q(or install if not exist): %v", ReleaseName, err)
	}
	return nil
}

// UpgradeOrInstall // 向 helm client 注册安装命令
func (k *KafkaManager) UpgradeOrInstall() error {
	chart := k.charts.Get(ChartName, "")
	if chart == nil {
		return fmt.Errorf("chart name=%q not exist", ChartName)
	}
	// values
	exporterResource := &v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
	if sr := k.spec.ExporterResources; sr != nil {
		if sr.Limits != nil {
			if sr.Limits.Cpu() != nil {
				exporterResource.Limits[v1.ResourceCPU] = *sr.Limits.Cpu()
			}
			if sr.Limits.Memory() != nil {
				exporterResource.Limits[v1.ResourceMemory] = *sr.Limits.Memory()
			}
		}
		if sr.Requests != nil {
			if sr.Requests.Cpu() != nil {
				exporterResource.Requests[v1.ResourceCPU] = *sr.Requests.Cpu()
			}
			if sr.Requests.Memory() != nil {
				exporterResource.Requests[v1.ResourceMemory] = *sr.Requests.Memory()
			}
		}
	}
	values := helm3.M{
		"namespace": ReleaseNamespace,
		"image": helm3.M{
			"registry": k.registry,
		},
		"enableDualStack": global.EnableDualStack,
		"depServices": helm3.M{
			"zookeeper": helm3.M{
				"host": "zookeeper-headless",
				"sasl": helm3.M{"enabled": true},
				"ssl":  helm3.M{"enabled": false},
			},
		},
		"config": helm3.M{
			"sasl": helm3.M{
				"enabled":   true,
				"mechanism": configuration.Plain,
				"user": helm3.L{
					helm3.M{
						"username": KafkaDefaultSSLUser,
						"password": KafkaDefaultSSLPassword,
					},
				},
			},
			"enableSSL": false,
			"kafkaENV": func() helm3.M {
				rel := make(helm3.M)
				for k, v := range k.spec.Env {
					rel[k] = v
				}
				return rel
			}(),
		},
		"service": helm3.M{"external": helm3.M{"enabled": true}},
		"replicaCount": func() int {
			replicas := len(k.spec.Hosts)
			if replicas == 0 {
				replicas = k.spec.ReplicaCount
			}
			return replicas
		}(),
		"resources": helm3.M{
			"exporter": exporterResource,
			"kafka":    k.spec.Resources.DeepCopy(),
		},
		"storage": helm3.M{
			"storageClassName": k.spec.StorageClassName,
			"local": func() helm3.M {
				rel := make(helm3.M)
				for i, host := range k.spec.Hosts {
					rel[strconv.Itoa(i)] = helm3.M{
						"host": host,
						"path": k.spec.Data_path,
					}
				}
				return rel
			}(),
		},
	}
	if len(k.spec.StorageCapacity) > 0 {
		values["storage"].(helm3.M)["capacity"] = k.spec.StorageCapacity
	}

	err := k.helm3.Upgrade(
		ReleaseName,
		helm3.ChartRefFromFile(filepath.Join(k.servicePackage, chart.Path)),
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValuesAny(values),
	)

	return err
}

func (k *KafkaManager) Reset() error {
	if !global.ClearData || k.spec.Data_path == "" {
		return nil
	}
	var wg sync.WaitGroup
	for _, node := range k.hosts {
		wg.Add(1)
		go func(host string, wg *sync.WaitGroup) {
			defer wg.Done()
			_ = universal.ClearDataDir(host, k.spec.Data_path)
		}(node.IP(), &wg)
	}
	wg.Wait()
	return nil
}
