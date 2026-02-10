package eceph

import (
	"strconv"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

const (
	// ProtonECephConfigManager defines variable used internally when referring to proton-eceph-config-manager component
	ProtonECephConfigManager = "proton-eceph-config-manager"
	// ProtonECephConfigWeb defines variable used internally when referring to proton-eceph-config-web component
	ProtonECephConfigWeb = "proton-eceph-config-web"
	// ProtonECephNodeAgent defines variable used internally when referring to proton-eceph-node-agent component
	ProtonECephNodeAgent = "proton-eceph-node-agent"
	// ProtonECephTenantWeb defines variable used internally when referring to proton-eceph-tenant-web component
	ProtonECephTenantWeb = "proton-eceph-tenant-web"
)

var KubernetesNamespace = configuration.GetProtonResourceNSFromFile()

var AllowedAccessLocalPorts = []string{
	strconv.Itoa(NGINXIngressControllerPortProtonECeph) + "/tcp",
	strconv.Itoa(NGINXIngressControllerPortProtonECephTenantWeb) + "/tcp",
	strconv.Itoa(NGINXServerPortECephHTTP) + "/tcp",
	strconv.Itoa(NGINXServerPortECephHTTPS) + "/tcp",
}

const (
	KeepalivedNameECephInternalVIP = "eceph-ivip"
	KeepalivedNameECephExternalVIP = "eceph_vip"
	KeepalivedNameECephSLB_HA      = "SLB_HA"

	KeepalivedVirtualRouterIDECephInternalVIP = 138
	KeepalivedVirtualRouterIDECephExternalVIP = 139
	KeepalivedVirtualRouterID_SLB_HA          = 112

	KeepalivedVirtualAddressLabelECephInternalVIP = "ivip"
	KeepalivedVirtualAddressLabelECephExternalVIP = "ovip"
	KeepalivedVirtualAddressLabelSLB_HA           = "ov"

	KeepalivedNotifyMasterECephExternalVIP = "/etc/slb/scripts/%s/entering_master.py"
	KeepalivedNotifyMasterECephSLB_HA      = "/etc/slb/scripts/SLB_HA/entering_master.py"

	KeepalivedNotifyBackupECephExternalVIP = "/etc/slb/scripts/%s/entering_backup.py"
	KeepalivedNotifyBackupECephSLB_HA      = "/etc/slb/scripts/SLB_HA/entering_backup.py"
)

const KubernetesServicePortConfigWeb = 14328

const (
	// Ceph RADOS (Reliable Autonomic Distributed Object Store) Gateway Port
	CephRADOSGatewayPort = 7480
)

const (
	NGINXServerNameECephHTTP  = "eceph_10001"
	NGINXServerNameECephHTTPS = "eceph_10002"

	NGINXServerPortECephHTTP  = 10001
	NGINXServerPortECephHTTPS = 10002

	NGINXServerCertificatePath = "/usr/local/slb-nginx/ssl/eceph-server.crt"
	NGINXServerKeyPath         = "/usr/local/slb-nginx/ssl/eceph-server.key"
)

const ECephDatabaseName = "minotaur"

const (
	NGINXIngressControllerPortProtonECeph          = 8003
	NGINXIngressControllerPortProtonECephTenantWeb = 8005

	IngressClassNGINX_ECeph              = "nginx-eceph"
	IngressClassNGINX_ECephConfigManager = "nginx-eceph-config-manage"
	IngressClassNGINX_ECephTenantWeb     = "nginx-eceph-tenant-web"
)

const ExecuteECephAgentConfigTimeout = "600s"

const SystemdUnitECephConfigAgent = "eceph-config-agent.service"

var SystemdUnits = []string{
	SystemdUnitECephConfigAgent,
}
