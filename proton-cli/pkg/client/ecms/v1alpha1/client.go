package v1alpha1

import (
	"net"
	"net/http"
	"net/url"
	"strconv"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/exec"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/files"
)

type Client struct {
	HTTPClient *http.Client

	Base *url.URL
}

// Create ecms/v1alpha1 client for hostname or ip address.
func NewForHost(host string) *Client {
	var rt http.RoundTripper = http.DefaultTransport
	rt = newDebuggingRoundTripper(rt)
	rt = newSimpleAuthRoundTripper(rt, "proton-cli")
	return &Client{
		HTTPClient: &http.Client{
			Transport: rt,
		},
		Base: &url.URL{
			Scheme: DefaultScheme,
			Host:   net.JoinHostPort(host, strconv.Itoa(DefaultPort)),
			Path:   DefaultAPIPath,
		},
	}
}

// Exec implements Interface.
func (c *Client) Exec() exec.Interface {
	return &exec.Client{
		HTTPClient: c.HTTPClient,
		Base:       c.Base.JoinPath("exec"),
	}
}

// Files implements Interface.
func (c *Client) Files() files.Interface {
	return &files.Client{
		HTTPClient: c.HTTPClient,
		Base:       c.Base.JoinPath("files"),
	}
}

var _ Interface = &Client{}
