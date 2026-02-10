package store

import (
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	node_v1alpha1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

type Manager struct {
	// Host of container registry.
	Registry string
	// deployment config of package store.
	Spec *configuration.PackageStore
	// package store's RDS client config.
	RDS *configuration.RdsInfo
	// client to access remote nodes which package store is running on.
	Nodes []node_v1alpha1.Interface
	// helm client to install/upgrade package store's helm release.
	Helm helm3.Client
	// 访问 Kubernetes 资源的客户端
	Kube client.Client

	ServicePackage *servicepackage.ServicePackage

	// easy for testing
	Logger logrus.FieldLogger

	Namespace string
}
