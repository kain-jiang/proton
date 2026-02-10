package grafana

import (
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

// Manager 负责安装、更新服务 Grafana，清理 Grafana 的数据目录
type Manager struct {
	// container image registry's address, host or host:port
	Registry string
	// grafana deployment configuration
	Spec *configuration.Grafana
	// 远程操访问节点的客户端
	Node v1alpha1.Interface
	// 调用 helm 的接口
	Helm helm3.Client

	ServicePackage *servicepackage.ServicePackage
	// prometheus's kubernetes service
	Prometheus *corev1.Service

	Logger logrus.FieldLogger

	Namespace string
}
