package registry

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/distribution/reference"
	"github.com/sirupsen/logrus"
)

// Interface 定义访问 container registry 的客户端
type Interface interface {
	// Address 返回 container registry 的地址，可能包含端口如果端口不是 80 或
	// 443
	Address() string

	// ListRepositoryTags 返回指定 repository 的 tags
	ListRepositoryTags(ctx context.Context, repository string) (tags []string, err error)
}

type Client struct {
	// container registry 的地址，可能包含端口
	address string

	sys *types.SystemContext

	log logrus.FieldLogger
}

func New(config *Config) (*Client, error) {
	if strings.Contains(config.Address, ":") {
		if _, _, err := net.SplitHostPort(config.Address); err != nil {
			return nil, err
		}
	}
	return &Client{address: config.Address, sys: &types.SystemContext{DockerInsecureSkipTLSVerify: types.OptionalBoolTrue}}, nil
}

// Address 返回 container registry 的地址，可能包含端口如果端口不是 80 或
// 443
func (c *Client) Address() string {
	return c.address
}

// ListRepositoryTags 返回指定 repository 的 tags
func (c *Client) ListRepositoryTags(ctx context.Context, repository string) ([]string, error) {
	named := c.address + "/" + repository
	ref, err := reference.ParseNormalizedNamed(named)
	if err != nil {
		return nil, fmt.Errorf("parse normalized named %q fail: %w", named, err)
	}

	imgRef, err := docker.NewReference(reference.TagNameOnly(ref))
	if err != nil {
		return nil, fmt.Errorf("create docker reference for named %q fail: %w", ref, err)
	}

	tags, err := docker.GetRepositoryTags(ctx, c.sys, imgRef)
	if err != nil {
		return nil, fmt.Errorf("get tags from docker repository %q fail: %w", imgRef, err)
	}

	return tags, nil
}

var _ Interface = (*Client)(nil)
