package fake

import (
	"context"
	"fmt"
	"net/http"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/registry"
)

// Client is a fake client for registry.Interface
type Client struct {
	address string

	repositories map[string][]string
}

func New(address string, repositories map[string][]string) *Client {
	return &Client{address: address, repositories: repositories}
}

// Address implements registry.Interface.
func (c *Client) Address() string {
	return c.address
}

// ListRepositoryTags implements registry.Interface.
func (c *Client) ListRepositoryTags(ctx context.Context, repository string) (tags []string, err error) {
	for r, ts := range c.repositories {
		if r == repository {
			tags = make([]string, len(ts))
			copy(tags, ts)
			return
		}
	}
	return nil, fmt.Errorf("list repository tags: invalid status code from registry %d (%s)", http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

var _ registry.Interface = (*Client)(nil)
