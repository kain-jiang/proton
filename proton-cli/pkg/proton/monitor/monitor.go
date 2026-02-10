package monitor

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

const (
	// Monitor 的 Chart 名称
	ChartName = "proton-monitor"
	// Monitor 的 Helm release 名称
	ReleaseName = "proton-monitor"
	// Monitor 的 Helm release 所在的命名空间
	ReleaseNamespace = metav1.NamespaceSystem
	// Monitor 的 etcd 证书 Secret 名称
	K8sEtcdCertsSecretName = "k8s-etcd-certs"
)

var log = logger.NewLogger()

type MonitorManager struct {
	// Monitor Spec
	spec *configuration.ProtonMonitor

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
	oldConf *configuration.ProtonMonitor
}

// NewManager 创建一个新的 MonitorManager 实例，用于在 apply.go 中调用
func NewManager(helm3 helm3.Client, spec *configuration.ProtonMonitor, registry string, servicePackage string, charts servicepackage.Charts, hosts []configuration.Node) *MonitorManager {
	return &MonitorManager{
		spec:           spec,
		helm3:          helm3,
		registry:       registry,
		servicePackage: servicePackage,
		charts:         charts,
		hosts:          hosts,
	}
}

// 设置 Helm 客户端
func (m *MonitorManager) Helm3(helm3 helm3.Client) *MonitorManager {
	m.helm3 = helm3
	return m
}

// 设置节点信息，用于通过 ssh 远程创建数据目录
func (m *MonitorManager) Hosts(hosts []configuration.Node) *MonitorManager {
	m.hosts = hosts
	return m
}

// 设置 Registry 地址
func (m *MonitorManager) Registry(registry string) *MonitorManager {
	m.registry = registry
	return m
}

// 设置 service-package 的路径
func (m *MonitorManager) ServicePackage(servicePackage string) *MonitorManager {
	m.servicePackage = servicePackage
	return m
}

// 设置 chart 列表
func (m *MonitorManager) Charts(charts servicepackage.Charts) *MonitorManager {
	m.charts = charts
	return m
}

// 设置 oldConfig 旧配置
func (m *MonitorManager) OldConfig(oldConf *configuration.ProtonMonitor) *MonitorManager {
	m.oldConf = oldConf
	return m
}

func (m *MonitorManager) Apply() error {
	var ctx = context.TODO()
	// 创建数据目录
	for _, host := range m.spec.Hosts {
		f := ecms.NewForHost(host).Files()

		if info, err := f.Stat(ctx, m.spec.DataPath); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			log.Printf("host[%s] create directory %s", host, m.spec.DataPath)
			if err := f.Create(ctx, m.spec.DataPath, true, nil); err != nil {
				return err
			}
		} else if !info.IsDir() {
			return fmt.Errorf("host[%s] %s is not a directory", host, m.spec.DataPath)
		}
	}

	// 确保 etcd 证书 Secret 存在
	if err := m.createEtcdCertsSecret(); err != nil {
		return fmt.Errorf("failed to create etcd certificates secret: %w", err)
	}

	// 向 helm client 注册安装命令
	return m.apply()
}

func (m *MonitorManager) apply() error {
	log.Infof("Applying release=%s chart=%s", ReleaseName, ChartName)

	if err := m.UpgradeOrInstall(); err != nil {
		return fmt.Errorf("unable to upgrade release %q (or install if not exist): %v", ReleaseName, err)
	}
	return nil
}

// UpgradeOrInstall // 向 helm client 注册安装命令
func (m *MonitorManager) UpgradeOrInstall() error {
	chart := m.charts.Get(ChartName, "")
	if chart == nil {
		return fmt.Errorf("chart name=%q not exist", ChartName)
	}

	// 使用指定的命名空间
	helm3Client := m.helm3.NameSpace(ReleaseNamespace)

	return helm3Client.Upgrade(
		ReleaseName,
		helm3.ChartRefFromFile(filepath.Join(m.servicePackage, chart.Path)),
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeRecreatePods(true),
		helm3.WithUpgradeCreateNamespace(true),
		// To convert k8s.io/api/core/v1.ResourceRequirements to
		// map[string]interface{}, use helm3.WithUpgradeValuesAny instead of
		// WithUpgradeValues.
		helm3.WithUpgradeValuesAny(HelmValuesFor(m.spec, m.registry)),
	)
}

func (m *MonitorManager) Reset() error {
	if !global.ClearData || m.spec.DataPath == "" {
		return nil
	}
	var wg sync.WaitGroup
	for _, node := range m.hosts {
		wg.Add(1)
		go func(host string, wg *sync.WaitGroup) {
			defer wg.Done()
			_ = universal.ClearDataDir(host, m.spec.DataPath)
		}(node.IP(), &wg)
	}
	wg.Wait()
	return nil
}

// createEtcdCertsSecret 创建包含 etcd 证书的 Secret
func (m *MonitorManager) createEtcdCertsSecret() error {
	log.Infof("Creating etcd certificates secret %s in %s namespace", K8sEtcdCertsSecretName, ReleaseNamespace)

	_, k8sClient := client.NewK8sClient()
	if k8sClient == nil {
		return client.ErrKubernetesClientSetNil
	}

	// 检查 Secret 是否已存在
	_, err := k8sClient.CoreV1().Secrets(ReleaseNamespace).Get(context.TODO(), K8sEtcdCertsSecretName, metav1.GetOptions{})
	if err == nil {
		log.Infof("Secret %s already exists in %s namespace", K8sEtcdCertsSecretName, ReleaseNamespace)
		return nil
	}

	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to check if secret %s exists: %w", K8sEtcdCertsSecretName, err)
	}

	// 从本地文件系统读取 etcd 证书
	caCertPath := "/etc/kubernetes/pki/etcd/ca.crt"
	healthcheckClientCertPath := "/etc/kubernetes/pki/etcd/healthcheck-client.crt"
	healthcheckClientKeyPath := "/etc/kubernetes/pki/etcd/healthcheck-client.key"

	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("failed to read ca.crt from %s: %w", caCertPath, err)
	}

	healthcheckClientCert, err := os.ReadFile(healthcheckClientCertPath)
	if err != nil {
		return fmt.Errorf("failed to read healthcheck-client.crt from %s: %w", healthcheckClientCertPath, err)
	}

	healthcheckClientKey, err := os.ReadFile(healthcheckClientKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read healthcheck-client.key from %s: %w", healthcheckClientKeyPath, err)
	}

	// 创建 Secret
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      K8sEtcdCertsSecretName,
			Namespace: ReleaseNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"ca.crt":                 caCert,
			"healthcheck-client.crt": healthcheckClientCert,
			"healthcheck-client.key": healthcheckClientKey,
		},
	}

	_, err = k8sClient.CoreV1().Secrets(ReleaseNamespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret %s: %w", K8sEtcdCertsSecretName, err)
	}

	log.Infof("Secret %s created successfully in %s namespace", K8sEtcdCertsSecretName, ReleaseNamespace)
	return nil
}
