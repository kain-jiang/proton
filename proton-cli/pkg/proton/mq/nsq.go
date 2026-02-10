package mq

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"sync"

	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

const (
	// nsq 的 Chart 名称
	ChartName = "proton-mq-nsq"
	// nsq 的 Helm release 名称，与 chart 名称一致
	ReleaseName = "proton-mq-nsq"

	// service 以前用的旧连接地址后缀
	NsqdServiceNameSuffix = "svc.cluster.local"

	NsqdServiceName       = "proton-mq-nsq-nsqd"
	NsqdServicePort       = 4151
	NsqlookuodServiceName = "proton-mq-nsq-nsqlookupd"
	NsqlookuodServicePort = 4161
)

var log = logger.NewLogger()

type MQManager struct {
	// redis Spec
	spec *configuration.ProtonDataConf

	// 节点访问配置，用于生成 SSH 客户端配置
	hosts []configuration.Node

	// registry 地址
	registry string

	// Helm client
	helm3 helm3.Client

	// service-package 的路径
	servicePackage string
	// chart 列表
	charts servicepackage.Charts

	// oldConfig 旧配置
	oldConf *configuration.ProtonDataConf

	// nsq 的 Helm release 所在的命名空间
	releaseNamespace string
}

// 创建 MQManager
func New(spec *configuration.ProtonDataConf) *MQManager {
	return &MQManager{
		spec: spec,
	}
}

func (m *MQManager) Helm3(helm3 helm3.Client) *MQManager {
	m.helm3 = helm3
	return m
}

// 设置节点信息，用于通过 ssh 远程创建数据目录
func (m *MQManager) Hosts(hosts []configuration.Node) *MQManager {
	m.hosts = hosts
	return m
}

// 设置 Registry 地址
func (m *MQManager) Registry(registry string) *MQManager {
	m.registry = registry
	return m
}

// 设置 service-package 的路径
func (m *MQManager) ServicePackage(servicePackage string) *MQManager {
	m.servicePackage = servicePackage
	return m
}

// 设置 chart 列表
func (m *MQManager) Charts(charts servicepackage.Charts) *MQManager {
	m.charts = charts
	return m
}

// 设置 oldConfig 旧配置
func (m *MQManager) OldConfig(oldConf *configuration.ProtonDataConf) *MQManager {
	m.oldConf = oldConf
	return m
}

func (m *MQManager) ReleaseNamespace(namespace string) *MQManager {
	m.releaseNamespace = namespace
	return m
}

func (m *MQManager) Apply() error {
	var ctx = context.TODO()
	// 创建数据目录
	for _, host := range m.spec.Hosts {
		f := ecms.NewForHost(host).Files()

		if info, err := f.Stat(ctx, m.spec.Data_path); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			log.Printf("host[%s] create directory %s", host, m.spec.Data_path)
			if err := f.Create(ctx, m.spec.Data_path, true, nil); err != nil {
				return err
			}
		} else if !info.IsDir() {
			return fmt.Errorf("host[%s] %s is not a directory", host, m.spec.Data_path)
		}
	}
	// 向 helm client 注册安装命令
	return m.apply()

}

func (m *MQManager) apply() error {
	log.Infof("Applying release=%s chart=%s", ReleaseName, ChartName)

	if err := m.UpgradeOrInstall(); err != nil {
		return fmt.Errorf("unable to upgrade release %q (or install if not exist): %v", ReleaseName, err)
	}
	return nil
}

// UpgradeOrInstall // 向 helm client 注册安装命令
func (m *MQManager) UpgradeOrInstall() error {
	chart := m.charts.Get(ChartName, "")
	if chart == nil {
		return fmt.Errorf("chart name=%q not exist", ChartName)
	}

	values := helm3.M{
		"namespace": m.releaseNamespace,
		"image": helm3.M{
			"registry": m.registry,
		},
		"replicaCount": func() int {
			replicas := len(m.spec.Hosts)
			if replicas == 0 {
				replicas = m.spec.ReplicaCount
			}
			return replicas
		}(),
		"env": helm3.M{
			"language": "en_US.UTF-8",
			"timezone": "Asia/Shanghai",
		},
		"service": helm3.M{
			"enableDualStack": global.EnableDualStack,
			"nsqd": helm3.M{
				"httpPort": 4151,
				"tcpPort":  4150,
			},
			"nsqlookupd": helm3.M{
				"httpPort":  4161,
				"tcpPort":   4160,
				"haEnabled": (len(m.spec.Hosts) > 1),
			},
			"nsqadmin": helm3.M{"enabled": false},
		},
		"storage": helm3.M{
			"storageClassName": m.spec.StorageClassName,
			"local": func() helm3.M {
				rel := make(helm3.M)
				for i, host := range m.spec.Hosts {
					rel[strconv.Itoa(i)] = helm3.M{
						"host": host,
						"path": m.spec.Data_path,
					}
				}
				return rel
			}(),
		},
	}
	if len(m.spec.StorageCapacity) > 0 {
		values["storage"].(helm3.M)["capacity"] = m.spec.StorageCapacity
	}
	if m.spec.Resources != nil {
		values["resources"] = m.spec.Resources
	}

	return m.helm3.Upgrade(
		ReleaseName,
		helm3.ChartRefFromFile(filepath.Join(m.servicePackage, chart.Path)),
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValues(values),
	)
}

func (m *MQManager) Reset() error {
	if !global.ClearData || m.spec.Data_path == "" {
		return nil
	}
	var wg sync.WaitGroup
	for _, node := range m.hosts {
		wg.Add(1)
		go func(host string, wg *sync.WaitGroup) {
			defer wg.Done()
			_ = universal.ClearDataDir(host, m.spec.Data_path)
		}(node.IP(), &wg)
	}
	wg.Wait()
	return nil
}
