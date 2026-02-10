package componentmanage

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/registry"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/componentmanage/packages"
)

type Upgrader struct {
	Package  *packages.ComponentPackage
	Registry string
	NewCfg   *configuration.ClusterConfig
}

func NewComponentApply(oldCfg, newCfg *configuration.ClusterConfig, rgr *registry.Client, pkgs *packages.ComponentPackage) (*Applier, error) {
	return &Applier{
		OldCfg:   oldCfg,
		NewCfg:   newCfg,
		charts:   pkgs.Charts,
		registry: rgr.Address(),

		onlyInitComponent: true,
		extraImages:       pkgs.Images,
	}, nil
}
