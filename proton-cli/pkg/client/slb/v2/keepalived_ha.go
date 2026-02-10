package v2

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rest"
)

// KeepalivedHAsGetter has a method to return a KeepalivedHAInterface. A group's
// client should implement this interface.
type KeepalivedHAsGetter interface {
	KeepalivedHAs() KeepalivedHAInterface
}

// KeepalivedHAInterface has methods to work with KeepalivedHA resources.
type KeepalivedHAInterface interface {
	Create(ctx context.Context, name string, kha *KeepalivedHA) error
	Update(ctx context.Context, name string, kha *KeepalivedHA) error
	Delete(ctx context.Context, name string) error
	Get(ctx context.Context, name string) (*KeepalivedHA, error)
	GetRaw(ctx context.Context, name string) (map[string]interface{}, error)
	List(ctx context.Context) ([]string, error)
}

// keepalivedHAs implements KeepalivedHAInterface
type keepalivedHAs struct {
	client rest.Interface
}

// newKeepalivedHAs returns a KeepalivedHAs
func newKeepalivedHAs(c *SLB_V2Client) *keepalivedHAs {
	return &keepalivedHAs{client: c.RESTClient()}
}

// Create implements KeepalivedHAInterface.
func (c *keepalivedHAs) Create(ctx context.Context, name string, kha *KeepalivedHA) error {
	var r struct {
		Conf struct {
			VRRPInstance map[string]KeepalivedHA `json:"vrrp_instance,omitempty"`
		} `json:"conf,omitempty"`
	}

	r.Conf.VRRPInstance = map[string]KeepalivedHA{name: *kha}

	return c.client.Post().
		Resource("keepalived/ha").
		Body(&r).
		Do(ctx).
		Error()
}

// Delete implements KeepalivedHAInterface.
func (*keepalivedHAs) Delete(ctx context.Context, name string) error {
	panic("unimplemented")
}

// Get implements KeepalivedHAInterface.
func (c *keepalivedHAs) Get(ctx context.Context, name string) (*KeepalivedHA, error) {
	var r struct {
		VRRPInstance map[string]KeepalivedHA `json:"vrrp_instance,omitempty"`
	}
	err := c.client.Get().
		Resource("keepalived/ha").
		Name(name).
		Do(ctx).
		Into(&r)
	kha := r.VRRPInstance[name]
	if err != nil {
		return nil, err
	}
	return &kha, err
}

func (c *keepalivedHAs) GetRaw(ctx context.Context, name string) (map[string]interface{}, error) {
	var r struct {
		VRRPInstance map[string]map[string]interface{} `json:"vrrp_instance,omitempty"`
	}
	err := c.client.Get().
		Resource("keepalived/ha").
		Name(name).
		Do(ctx).
		Into(&r)
	kha := r.VRRPInstance[name]
	if err != nil {
		return nil, err
	}
	return kha, err
}

// List implements KeepalivedHAInterface.
func (c *keepalivedHAs) List(ctx context.Context) (result []string, err error) {
	err = c.client.Get().
		Resource("keepalived/ha").
		Do(ctx).
		Into(&result)
	return
}

// Update implements KeepalivedHAInterface.
func (c *keepalivedHAs) Update(ctx context.Context, name string, kha *KeepalivedHA) error {
	actual, err := c.GetRaw(context.TODO(), name)
	if err != nil {
		return err
	}

	// update the instance
	actual["interface"] = kha.Interface
	actual["priority"] = kha.Priority
	actual["unicast_peer"] = kha.UnicastPeer
	actual["unicast_src_ip"] = kha.UnicastSRC_IP
	actual["virtual_ipaddress"] = kha.VirtualIPAddress
	actual["virtual_router_id"] = kha.VirtualRouterID
	if len(kha.NotifyMaster) > 0 {
		actual["notify_master"] = kha.NotifyMaster
	}
	if len(kha.NotifyBackup) > 0 {
		actual["notify_backup"] = kha.NotifyBackup
	}

	var r struct {
		Conf struct {
			VRRPInstance map[string]map[string]interface{} `json:"vrrp_instance,omitempty"`
		} `json:"conf,omitempty"`
	}

	r.Conf.VRRPInstance = map[string]map[string]interface{}{name: actual}

	return c.client.Put().
		Resource("keepalived/ha").
		Name(name).
		Body(&r).
		Do(ctx).
		Error()
}

var _ KeepalivedHAInterface = &keepalivedHAs{}
