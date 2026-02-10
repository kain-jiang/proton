package v2

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rest"
)

type SLB_V2Interface interface {
	RESTClient() rest.Interface
	KeepalivedHAsGetter
}

// SLB_V2Client is used to interact with features provided by the slb group.
type SLB_V2Client struct {
	restClient rest.Interface
}

func (c *SLB_V2Client) KeepalivedHAs() KeepalivedHAInterface {
	return newKeepalivedHAs(c)
}

// NewForConfig creates a new SLB_V2Client for the given config. NewForConfig is
// equivalent to NewForConfigAndClient(c, httpClient), where httpClient was
// generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*SLB_V2Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	return &SLB_V2Client{client}, err
}

func setConfigDefaults(config *rest.Config) error {
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/api"
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultProtonCLIUserAgent()
	}
	return nil
}

// RESTClient returns a RESTClient that is used to communicate with API server
// by this client implementation.
func (c *SLB_V2Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
