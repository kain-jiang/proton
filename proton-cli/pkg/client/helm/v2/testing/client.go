package testing

import (
	"context"

	helm "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm/v2"
)

type Client struct {
	Err error
}

// UpdateRepoCache implements v2.Interface.
func (c *Client) UpdateRepoCache(ctx context.Context, repo string) error {
	return c.Err
}

// Reconcile implements v2.Interface.
func (c *Client) Reconcile(ctx context.Context, release string, chart string, values map[string]any) error {
	return c.Err
}

var _ helm.Interface = &Client{}
