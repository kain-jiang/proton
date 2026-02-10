package v1

import (
	"context"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rest"
)

type NginxHTTPGetter interface {
	NginxHTTPs() NginxHTTPInterface
}

type NginxHTTPInterface interface {
	Create(ctx context.Context, nh *NginxHTTP) error
	Update(ctx context.Context, nh *NginxHTTP) error
	Delete(ctx context.Context, name string) error
	Get(ctx context.Context, name string) (*NginxHTTP, error)
	List(ctx context.Context) ([]string, error)
}

type nginxHTTPs struct {
	client rest.Interface
}

func newNginxHTTPs(c *SLB_V1Client) *nginxHTTPs {
	return &nginxHTTPs{client: c.RESTClient()}
}

// Create implements NginxHTTPInterface.
func (c *nginxHTTPs) Create(ctx context.Context, nh *NginxHTTP) error {
	res := c.client.Post().
		Resource("nginx/nginx").
		Body(map[string]interface{}{
			"conf": map[string]string{
				"worker_processes": "auto",
			},
		}).
		Do(ctx)
	// ignore 409 -- that means nginx instance we are going to create already exists
	if res.Error() != nil && res.StatusCode() != 409 {
		return res.Error()
	}
	return c.client.Post().
		Resource("nginx/http").
		Body(nh).
		Do(ctx).
		Error()
}

// Delete implements NginxHTTPInterface.
func (c *nginxHTTPs) Delete(ctx context.Context, name string) error {
	return c.client.Delete().
		Resource("nginx/http").
		Name(name).
		Do(ctx).
		Error()
}

// Get implements NginxHTTPInterface.
func (c *nginxHTTPs) Get(ctx context.Context, name string) (result *NginxHTTP, err error) {
	result = &NginxHTTP{}
	err = c.client.Get().
		Resource("nginx/http").
		Name(name).
		Do(ctx).
		Into(result)
	return
}

// List implements NginxHTTPInterface.
func (c *nginxHTTPs) List(ctx context.Context) (result []string, err error) {
	err = c.client.Get().
		Resource("nginx/http").
		Do(ctx).
		Into(&result)
	return
}

// Update implements NginxHTTPInterface.
func (c *nginxHTTPs) Update(ctx context.Context, nh *NginxHTTP) error {
	return c.client.Put().
		Resource("nginx/http").
		Name(nh.Name).
		Body(nh).
		Do(ctx).
		Error()
}

var _ NginxHTTPInterface = &nginxHTTPs{}
