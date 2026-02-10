package configuration

import (
	"os"
)

func GetProtonDeployConf() string {
	b, _ := os.ReadFile("/etc/cluster.yaml")
	return string(b)
}

// proton-deploy cluster conf
type ProtonDeployClusterConfig map[string]interface{}

const (
	// max host count allowed in proton-deploy conf file
	ProtonDeployMaxHostsAllowed int = 255
	// max slb ha allowed in proton-deploy conf file, currently proton-cli is not designed to handle anything more than 1
	ProtonDeployMaxSLBHAsAllowed int = 1
	// top level field names
	FnProtonDeployAPIVersion string = "apiVersion"
	FnProtonDeployHosts      string = "hosts"
	FnProtonDeploySLB        string = "slb"
	FnProtonDeployCS         string = "cs"
	FnProtonDeployECeph      string = "eceph"
	// lower level field names
	FnProtonDeployHostsSSHIP       string = "ssh_ip"
	FnPDHostsInternalIP            string = "internal_ip"
	FnProtonDeploySLBListenPort    string = "slb_listen"
	FnProtonDeployHighlyAvailable  string = "ha"
	FnProtonDeployHALabel          string = "label"
	FnProtonDeployHAVirtualIP      string = "vip"
	FnProtonDeployECephHosts       string = "hosts"
	FnProtonDeployECephNamespace   string = "namespace"
	FnProtonDeployECephSelfCert    string = "self_cert"
	FnPDECephSrcSSLPath            string = "src_ssl_path"
	FnPDECephDstSSLPath            string = "dst_ssl_path"
	FnPDECephCertFile              string = "crt_file"
	FnPDECephKeyFile               string = "key_file"
	FnPDECephLoadBalancer          string = "lb"
	FnPDECephLoadBalancerVirtualIP string = "vip"
	// default values
	PDDefaultECephSelfCert  bool   = false
	PDDefaultDstSSLPath     string = "/usr/local/slb-nginx/ssl"
	PDDefaultCrtFileName    string = "eceph-server.crt"
	PDDefaultKeyFileName    string = "eceph-server.key"
	PDDefaultECephNamespace string = "default"
)
