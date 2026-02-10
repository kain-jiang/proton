package v1alpha1

import (
	"net"

	eceph_agent_config "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/eceph/agent_config/v1alpha1"
	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	firewalld "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/firewalld/v1alpha1"
	helm "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm/v2"
	slb_v1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1"
	slb_v2 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
	systemd "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/systemd/v1alpha1"
)

// Interface 的是访问节点的客户端的接口定义，用于读写节点的文件，执行命令
type Interface interface {
	// Name 返回客户端所对应节点的名称
	Name() string

	IP() net.IP

	IPVersion() string

	InternalIP() net.IP

	NetworkInterfaces() ([]NetworkInterface, error)

	// Node systemd interface
	Systemd() systemd.Interface

	// Node firewalld v1alpha1
	Firewalld() firewalld.Interface

	ECMS() ecms.Interface

	// Proton ECeph agent_config v1alpha1
	ECephAgentConfig() eceph_agent_config.Interface

	// Proton SLB v1
	SLB_V1() slb_v1.SLB_V1Interface
	// Proton SLB v2
	SLB_V2() slb_v2.SLB_V2Interface

	// Deprecated: use helm/v3 instead
	Helm() helm.Interface
}
