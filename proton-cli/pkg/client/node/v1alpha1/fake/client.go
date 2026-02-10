package fake

import (
	"net"
	"testing"

	eceph_agent_config "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/eceph/agent_config/v1alpha1"
	eceph_agent_config_testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/eceph/agent_config/v1alpha1/testing"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	firewalld "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/firewalld/v1alpha1"
	firewalld_testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/firewalld/v1alpha1/testing"
	helm "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm/v2"
	helm_testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm/v2/testing"
	node "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	slb_v1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1"
	slb_v1_testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1/testing"
	slb_v2 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
	slb_v2_testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2/testing"
	systemd "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/systemd/v1alpha1"
	systemd_testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/systemd/v1alpha1/testing"
)

// Fake node.v1alpha1.Interface
type Client struct {
	name string

	ipv4, ipv6, internal net.IP
}

func New(name string) *Client { return &Client{name: name} }

// NewForTesting return Client with given directories and files
//
//	node := NewForTesting(t, "node-x", []string{"/path/to/directory"}, []string{"/path/to/file"})
func NewForTesting(t *testing.T, name string, directories, files []string) *Client {
	return &Client{name: name}
}

func (c *Client) Name() string { return c.name }

// IP implements v1alpha1.Interface.
func (c *Client) IP() net.IP {
	if c.ipv4 != nil {
		return c.ipv4
	}
	return c.ipv6
}

// IPVersion return the type of returned IP by IP()
func (c *Client) IPVersion() string {
	if c.ipv4 != nil {
		return "ipv4"
	}
	return "ipv6"
}

// InternalIP implements v1alpha1.Interface.
func (c *Client) InternalIP() net.IP {
	return c.internal
}

// NetworkInterfaces implements v1alpha1.Interface.
func (c *Client) NetworkInterfaces() (interfaces []node.NetworkInterface, err error) {
	var external node.NetworkInterface
	if c.ipv4 != nil {
		external.Addresses = append(external.Addresses, net.IPNet{IP: c.ipv4, Mask: net.CIDRMask(24, 32)})
	}
	if c.ipv6 != nil {
		external.Addresses = append(external.Addresses, net.IPNet{IP: c.ipv6, Mask: net.CIDRMask(64, 128)})
	}
	interfaces = append(interfaces, external)

	if c.internal == nil {
		return
	}

	var mask net.IPMask
	if c.internal.To4() == nil {
		mask = net.CIDRMask(64, 128)
	} else {
		mask = net.CIDRMask(24, 32)
	}

	var internal node.NetworkInterface
	internal.Addresses = append(internal.Addresses, net.IPNet{IP: c.internal, Mask: mask})
	return
}

// Systemd implements v1alpha1.Interface.
func (*Client) Systemd() systemd.Interface {
	return new(systemd_testing.Client)
}

// Firewalld implements v1alpha1.Interface.
func (*Client) Firewalld() firewalld.Interface {
	return new(firewalld_testing.Client)
}

// ECMS implements v1alpha1.Interface.
func (c *Client) ECMS() v1alpha1.Interface {
	panic("unimplemented")
}

// ECephAgentConfig implements v1alpha1.Interface.
func (*Client) ECephAgentConfig() eceph_agent_config.Interface {
	return new(eceph_agent_config_testing.Client)
}

// SLB_V1 implements v1alpha1.Interface.
func (*Client) SLB_V1() slb_v1.SLB_V1Interface {
	return new(slb_v1_testing.Client)
}

// SLB_V2 implements v1alpha1.Interface.
func (*Client) SLB_V2() slb_v2.SLB_V2Interface {
	return new(slb_v2_testing.Client)
}

// Deprecated: use helm/v3 instead
func (*Client) Helm() helm.Interface {
	return new(helm_testing.Client)
}

var _ node.Interface = &Client{}
