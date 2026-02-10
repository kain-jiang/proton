package zookeeper

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
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
	// zookeeper 的 Chart 名称
	ChartName = configuration.ChartNameZooKeeper
	// zookeeper 的 Helm release 名称
	ReleaseName = configuration.ReleaseNameZooKeeper
	// zookeeper 的 Helm release 所在的命名空间
	ReleaseNamespace = "resource"
)

var log = logger.NewLogger()

type ZookeeperManager struct {
	// Kookeeper Spec
	spec *configuration.ZooKeeper

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
	oldConf *configuration.ZooKeeper
}

// 创建 ZookeeperManager
func New(spec *configuration.ZooKeeper) *ZookeeperManager {
	return &ZookeeperManager{
		spec: spec,
	}
}

// 设置 Helm 客户端
func (z *ZookeeperManager) Helm3(helm3 helm3.Client) *ZookeeperManager {
	z.helm3 = helm3
	return z
}

// 设置节点信息，用于通过 ssh 远程创建数据目录
func (z *ZookeeperManager) Hosts(hosts []configuration.Node) *ZookeeperManager {
	z.hosts = hosts
	return z
}

// 设置 Registry 地址
func (z *ZookeeperManager) Registry(registry string) *ZookeeperManager {
	z.registry = registry
	return z
}

// 设置 service-package 的路径
func (z *ZookeeperManager) ServicePackage(servicePackage string) *ZookeeperManager {
	z.servicePackage = servicePackage
	return z
}

// 设置 chart 列表
func (z *ZookeeperManager) Charts(charts servicepackage.Charts) *ZookeeperManager {
	z.charts = charts
	return z
}

// 设置 oldConfig 旧配置
func (z *ZookeeperManager) OldConfig(oldConf *configuration.ZooKeeper) *ZookeeperManager {
	z.oldConf = oldConf
	return z
}

func (z *ZookeeperManager) Apply() error {
	var ctx = context.TODO()
	// 创建数据目录
	for _, host := range z.spec.Hosts {
		f := ecms.NewForHost(host).Files()

		if info, err := f.Stat(ctx, z.spec.Data_path); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			log.Printf("host[%s] create directory %s", host, z.spec.Data_path)
			if err := f.Create(ctx, z.spec.Data_path, true, nil); err != nil {
				return err
			}
		} else if !info.IsDir() {
			return fmt.Errorf("host[%s] %s is not a directory", host, z.spec.Data_path)
		}
	}
	// 向 helm client 注册安装命令
	return z.apply()

}

func (z *ZookeeperManager) apply() error {
	log.Infof("Applying release=%s chart=%s", ReleaseName, ChartName)

	if err := z.UpgradeOrInstall(); err != nil {
		return fmt.Errorf("unable to upgrade release %q (or install if not exist): %v", ReleaseName, err)
	}
	return nil
}

// UpgradeOrInstall // 向 helm client 注册安装命令
func (z *ZookeeperManager) UpgradeOrInstall() error {
	chart := z.charts.Get(ChartName, "")
	if chart == nil {
		return fmt.Errorf("chart name=%q not exist", ChartName)
	}

	return z.helm3.Upgrade(
		ReleaseName,
		helm3.ChartRefFromFile(filepath.Join(z.servicePackage, chart.Path)),
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeRecreatePods(true),
		// To convert k8s.io/api/core/v1.ResourceRequirements to
		// map[string]interface{}, use helm3.WithUpgradeValuesAny instead of
		// WithUpgradeValues.
		//
		// issue: https://devops.aishu.cn/AISHUDevOps/ICT/_workitems/edit/459711/
		helm3.WithUpgradeValuesAny(HelmValuesFor(z.spec, z.registry)),
	)
}

func (m *ZookeeperManager) Reset() error {
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
