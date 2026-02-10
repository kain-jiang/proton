package testing

import (
	"k8s.io/utils/strings/slices"

	firewalld "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/firewalld/v1alpha1"
)

type Client struct {
	Running bool

	// permanent - zone - ports
	// 	ports := c.Ports[permanent][zone]
	Ports map[bool]map[string][]string

	Err error
}

// ChangeZoneInterfaces implements v1alpha1.Interface.
func (c *Client) ChangeZoneInterfaces(permanent bool, zone string, interfaces []string) error {
	panic("unimplemented")
}

// DeleteIPSet implements v1alpha1.Interface.
func (c *Client) DeleteIPSet(ipset string) error {
	panic("unimplemented")
}

// DeleteZone implements v1alpha1.Interface.
func (c *Client) DeleteZone(zone string) error {
	panic("unimplemented")
}

// GetZoneTarget implements v1alpha1.Interface.
func (c *Client) GetZoneTarget(zone string) (firewalld.Target, error) {
	panic("unimplemented")
}

// SetZoneTarget implements v1alpha1.Interface.
func (c *Client) SetZoneTarget(zone string, target firewalld.Target) error {
	panic("unimplemented")
}

// AddIPSetEntries implements v1alpha1.Interface.
func (c *Client) AddIPSetEntries(permanent bool, ipset string, entries []string) error {
	panic("unimplemented")
}

// AddZoneInterfaces implements v1alpha1.Interface.
func (c *Client) AddZoneInterfaces(permanent bool, zone string, interfaces []string) error {
	panic("unimplemented")
}

// AddZoneSources implements v1alpha1.Interface.
func (c *Client) AddZoneSources(permanent bool, zone string, sources []string) error {
	panic("unimplemented")
}

// GetIPSetEntries implements v1alpha1.Interface.
func (c *Client) GetIPSetEntries(permanent bool, ipset string) ([]string, error) {
	panic("unimplemented")
}

// GetIPSets implements v1alpha1.Interface.
func (c *Client) GetIPSets(permanent bool) ([]string, error) {
	panic("unimplemented")
}

// GetTarget implements v1alpha1.Interface.
func (c *Client) GetTarget(zone string) (firewalld.Target, error) {
	panic("unimplemented")
}

// GetZones implements v1alpha1.Interface.
func (c *Client) GetZones(permanent bool) ([]string, error) {
	panic("unimplemented")
}

// ListZoneInterfaces implements v1alpha1.Interface.
func (c *Client) ListZoneInterfaces(permanent bool, zone string) ([]string, error) {
	panic("unimplemented")
}

// ListZoneSources implements v1alpha1.Interface.
func (c *Client) ListZoneSources(permanent bool, zone string) ([]string, error) {
	panic("unimplemented")
}

// NewIPSet implements v1alpha1.Interface.
func (c *Client) NewIPSet(ipset string, ipsetType firewalld.IPSetType, family firewalld.IPSetFamily) error {
	panic("unimplemented")
}

// NewZone implements v1alpha1.Interface.
func (c *Client) NewZone(zone string) error {
	panic("unimplemented")
}

// Reload implements v1alpha1.Interface.
func (c *Client) Reload() error {
	panic("unimplemented")
}

// RemoveIPSetEntries implements v1alpha1.Interface.
func (c *Client) RemoveIPSetEntries(permanent bool, ipset string, entries []string) error {
	panic("unimplemented")
}

// RemoveZoneInterfaces implements v1alpha1.Interface.
func (c *Client) RemoveZoneInterfaces(permanent bool, zone string, interfaces []string) error {
	panic("unimplemented")
}

// RemoveZoneSources implements v1alpha1.Interface.
func (c *Client) RemoveZoneSources(permanent bool, zone string, sources []string) error {
	panic("unimplemented")
}

// SetTarget implements v1alpha1.Interface.
func (c *Client) SetTarget(zone string, target firewalld.Target) error {
	panic("unimplemented")
}

// State implements v1alpha1.Interface.
func (c *Client) State() (bool, error) {
	panic("unimplemented")
}

// stub
func (c *Client) RemoveAndAddFirewallSource(ip string, oldIP string, permanent bool) error {
	return nil
}

// AddPort implements v1alpha1.Interface.
func (c *Client) AddPort(port string, permanent bool, zone string, timeout string) error {
	if c.Err != nil {
		return c.Err
	}

	if c.Ports == nil {
		c.Ports = make(map[bool]map[string][]string)
	}

	if c.Ports[permanent] == nil {
		c.Ports[permanent] = make(map[string][]string)
	}

	if slices.Contains(c.Ports[permanent][zone], port) {
		return nil
	}

	c.Ports[permanent][zone] = append(c.Ports[permanent][zone], port)
	return nil
}

// IsRunning implements v1alpha1.Interface.
func (c *Client) IsRunning() (bool, error) {
	if c.Err != nil {
		return false, c.Err
	}
	return c.Running, nil
}

// ListPorts implements v1alpha1.Interface.
func (c *Client) ListPorts(permanent bool, zone string) (ports []string, err error) {
	if c.Err != nil {
		return nil, c.Err
	}

	if c.Ports != nil && c.Ports[permanent] != nil {
		ports = make([]string, len(c.Ports[permanent][zone]))
		copy(ports, c.Ports[permanent][zone])
	}

	return
}

var _ firewalld.Interface = &Client{}
