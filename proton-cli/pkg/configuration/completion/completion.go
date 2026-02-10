package completion

import (
	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	eceph "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/eceph/completion"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

// CompleteClusterConfig 补全 ClusterConfig
func CompleteClusterConfig(c, oldc *configuration.ClusterConfig, pkg *servicepackage.ServicePackage) {
	CompletionCS(c.Cs, c.Nodes)
	c.Chrony = CompleteChrony(c.Chrony, c.Cs)
	CompleteFirewall(&c.Firewall)
	if c.Proton_mq_nsq != nil {
		CompleteNSQ(c.Proton_mq_nsq, pkg.Charts())
	}
	if c.Prometheus != nil {
		CompletePrometheus(c.Prometheus)
	}
	if c.Grafana != nil {
		CompleteGrafana(c.Grafana)
	}

	CompleteInternalInfo(c)
	if c.PackageStore != nil {
		CompletePackageStore(c.PackageStore)
	}
}

// do completions that requires clients like node client here
func CompleteClusterConfigPost(c *configuration.ClusterConfig, pkg *servicepackage.ServicePackage, n []node.Interface) {
	if c.ECeph != nil {
		eceph.CompletePost(c.ECeph, c.Nodes, n)
	}
}
