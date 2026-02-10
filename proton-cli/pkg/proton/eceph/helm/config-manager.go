package helm

import "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"

func ValuesForConfigManager(registry string, rds *configuration.RdsInfo, rgwHost string, rgwPort int, rgwProtocol string, bindAddr string) *Values4ECeph {
	return &Values4ECeph{
		Namespace: NamespaceDefault,
		Image: ValuesImage{
			Registry: registry,
		},
		DepServices: &ValuesDepServices{
			RDS: &ValuesRDS{
				Type:     string(rds.RdsType),
				Host:     rds.Hosts,
				Port:     rds.Port,
				User:     rds.Username,
				Password: rds.Password,
			},
			RGW: &ValuesRGW{
				Host:     rgwHost,
				Port:     rgwPort,
				Protocol: rgwProtocol,
			},
		},
		Service: map[string]string{
			"bind-addr": bindAddr,
		},
	}
}
