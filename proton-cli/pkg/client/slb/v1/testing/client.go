package testing

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rest"
	slb "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v1"
)

type Client struct{}

// NginxHTTPs implements v1.SLB_V1Interface.
func (*Client) NginxHTTPs() slb.NginxHTTPInterface {
	return new(NGINXHttp)
}

// RESTClient implements v1.SLB_V1Interface.
func (*Client) RESTClient() rest.Interface {
	panic("unimplemented")
}

var _ slb.SLB_V1Interface = &Client{}
