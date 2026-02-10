package helm

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NamespaceAnyShare = "anyshare"
	NamespaceDefault  = meta_v1.NamespaceDefault
)

const (
	ChartNameConfigManager          = "proton-eceph-config-manager"
	ChartNameConfigWeb              = "proton-eceph-config-web"
	ChartNameNodeAgent              = "proton-eceph-node-agent"
	ChartNameTenantWeb              = "proton-eceph-tenant-web"
	ChartNameNGINXIngressController = "nginx-ingress-controller"

	ReleaseNameConfigManager        = ChartNameConfigManager
	ReleaseNameConfigWeb            = ChartNameConfigWeb
	ReleaseNameNodeAgent            = ChartNameNodeAgent
	ReleaseNameTenantWeb            = ChartNameTenantWeb
	ReleaseNameNGINX_ECeph          = "nginx-eceph"
	ReleaseNameNGINX_ECephTenantWeb = "nginx-eceph-tenant-web"
)
