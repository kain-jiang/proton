package v1

import "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rest"

type SLB_V1Interface interface {
	RESTClient() rest.Interface
	NginxHTTPGetter
}

type SLB_V1Client struct {
	restClient rest.Interface
}

func (c *SLB_V1Client) NginxHTTPs() NginxHTTPInterface {
	return newNginxHTTPs(c)
}

// NewForConfig creates a new SLB_V1Client for the given config. NewForConfig is
// equivalent to NewForConfigAndClient(c, httpClient), where httpClient was
// generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*SLB_V1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	return &SLB_V1Client{client}, err
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
func (c *SLB_V1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
