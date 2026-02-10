package testing

import (
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rest"
	slb "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
)

type Client struct{}

// KeepalivedHAs implements v2.SLB_V2Interface.
func (*Client) KeepalivedHAs() slb.KeepalivedHAInterface {
	return new(KeepalivedHA)
}

// RESTClient implements v2.SLB_V2Interface.
func (*Client) RESTClient() rest.Interface {
	panic("unimplemented")
}

var _ slb.SLB_V2Interface = &Client{}
