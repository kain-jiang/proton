package testing

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/exp/slices"

	slb "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1"
)

type NGINXHttp struct {
	Items []slb.NginxHTTP

	Err error
}

// Create implements v1.NginxHTTPInterface.
func (c *NGINXHttp) Create(ctx context.Context, nh *slb.NginxHTTP) error {
	if c.Err != nil {
		return c.Err
	}

	index, found := slices.BinarySearchFunc(c.Items, nh.Name, func(nh slb.NginxHTTP, s string) int { return strings.Compare(nh.Name, s) })
	if found {
		return errors.New("already exists")
	}

	c.Items = slices.Insert[[]slb.NginxHTTP, slb.NginxHTTP](c.Items, index, *nh)
	return nil
}

// Delete implements v1.NginxHTTPInterface.
func (c *NGINXHttp) Delete(ctx context.Context, name string) error {
	if c.Err != nil {
		return c.Err
	}
	index, found := slices.BinarySearchFunc(c.Items, name, func(nh slb.NginxHTTP, s string) int { return strings.Compare(nh.Name, s) })
	if !found {
		return errors.New("not found")
	}
	c.Items = slices.Delete[[]slb.NginxHTTP, slb.NginxHTTP](c.Items, index, index)
	return nil
}

// Get implements v1.NginxHTTPInterface.
func (c *NGINXHttp) Get(ctx context.Context, name string) (*slb.NginxHTTP, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	index, found := slices.BinarySearchFunc(c.Items, name, func(nh slb.NginxHTTP, s string) int { return strings.Compare(nh.Name, s) })
	if !found {
		return nil, errors.New("not found")
	}

	return &c.Items[index], nil
}

// List implements v1.NginxHTTPInterface.
func (c *NGINXHttp) List(ctx context.Context) (result []string, err error) {
	if c.Err != nil {
		return nil, c.Err
	}
	for _, item := range c.Items {
		result = append(result, item.Name)
	}
	return
}

// Update implements v1.NginxHTTPInterface.
func (c *NGINXHttp) Update(ctx context.Context, nh *slb.NginxHTTP) error {
	if c.Err != nil {
		return c.Err
	}
	index, found := slices.BinarySearchFunc(c.Items, nh.Name, func(nh slb.NginxHTTP, s string) int { return strings.Compare(nh.Name, s) })
	if !found {
		return errors.New("not found")
	}
	c.Items[index] = *nh
	return nil
}

var _ slb.NginxHTTPInterface = &NGINXHttp{}
