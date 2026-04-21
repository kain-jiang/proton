package v2

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rest"
)

type keepalivedConf struct {
	client rest.Interface
}

func NewKeepalivedConf(c *SLB_V2Client) *keepalivedConf {
	return &keepalivedConf{client: c.RESTClient()}
}

func (c *keepalivedConf) Get(ctx context.Context) (map[string]any, error) {
	var result map[string]any
	err := c.client.Get().
		Resource("keepalived/keepalived").
		Do(ctx).
		Into(&result)
	if err != nil {
		return nil, err
	}
	// convert back KeepalivedHAResponse type for response to general KeepalivedHAtype
	return result, err
}

// Update implements KeepalivedHAInterface.
func (c *keepalivedConf) Update(ctx context.Context, conf map[string]any) error {
	return c.client.Put().
		Resource("keepalived/keepalived").
		Body(&conf).
		Do(ctx).
		Error()
}
