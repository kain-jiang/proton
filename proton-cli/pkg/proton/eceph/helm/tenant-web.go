package helm

func ValuesForTenantWeb(registry string, rgwHost string, rgwPort int, rgwProtocol string, bindAddr string, tenantMgrHost string) *Values4ECeph {
	return &Values4ECeph{
		Namespace: NamespaceDefault,
		Image: ValuesImage{
			Registry: registry,
		},
		DepServices: &ValuesDepServices{
			RGW: &ValuesRGW{
				Host:     rgwHost,
				Port:     rgwPort,
				Protocol: rgwProtocol,
			},
			ProtonECephTenantManager: map[string]string{
				"host": tenantMgrHost,
			},
		},
		Service: map[string]string{
			"bind-addr": bindAddr,
		},
	}
}
