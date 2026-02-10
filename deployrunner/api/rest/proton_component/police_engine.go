package proton_component

import (
	"encoding/json"
	"net/http"

	"component-manage/pkg/models/types"
	"component-manage/pkg/models/types/components"
	"taskrunner/api/rest/proton_component/cmp"

	"github.com/gin-gonic/gin"
)

// GetProtonPolicyEngine get policyengine
//
//	@Summary	获取proton的policyengine实例
//
//	@Tags		proton,ProtonCompoment,policyengine
//	@Accept		json
//	@Produce	json
//
//	@Param		name									path		string						true	"实例名称"
//
//	@Success	200										{object}	types.ComponentPolicyEngine	"实例数据"
//	@Failure	400										{object}	HTTPError					"客户端请求参数错误"
//	@Failure	404										{object}	HTTPError					"对象不存在"
//	@Failure	500										{object}	HTTPError					"系统内部错误"
//	@Router		/components/release/policyengine/{name}	[get]
func (s *Server) GetProtonPolicyEngine(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("must set zookeeper realse name").abortGin(ctx)
	// 	return
	// }
	name := pleRlsName

	rls := &cmp.ComponentInstance[types.ComponentPolicyEngine]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "policyengine",
		},
	}
	GetProtonReleaseMarco(ctx, s, rls)
	if !ctx.IsAborted() {
		ctx.JSON(http.StatusOK, rls.Instance)
	}
}

// UpdateProtonPolicyEngine update proton policyengine instance
//
//	@Summary		更新proton policyengine
//	@Description	更新proton policyengine
//
//	@Tags			proton,ProtonCompoment,policyengine
//	@Accept			json
//	@Produce		json
//
//	@Param			obj									body		types.ComponentPolicyEngine	true	"zookeeper配置信息"
//
//	@Success		200									{object}	int							"返回为空"
//	@Failure		400									{object}	HTTPError					"客户端请求参数错误"
//	@Failure		404									{object}	HTTPError					"对象不存在，不允许更新"
//	@Failure		500									{object}	HTTPError					"系统内部错误"
//	@Router			/components/release/policyengine 	[PUT]
func (s *Server) UpdateProtonPolicyEngine(ctx *gin.Context) {
	obj := &types.ComponentPolicyEngine{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	// if obj.Name == "" {
	// 	paramError.From("name must be not empty").abortGin(ctx)
	// 	return
	// }
	obj.Name = pleRlsName

	// if obj.Dependencies.ETCD == ""{
	// 	obj.Dependencies.ETCD = etcdRlsName
	// }
	if obj.Dependencies == nil || obj.Dependencies.ETCD == "" {
		obj.Dependencies = &components.PolicyEngineComponentDependencies{
			ETCD: etcdRlsName,
		}
	}

	rls := cmp.ComponentInstance[types.ComponentPolicyEngine]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "policyengine",
		},
		Instance: *obj,
	}

	UpdateProtonReleaseMarco(ctx, s, rls)
	if ctx.IsAborted() {
		return
	}

	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	conf.Proton_policy_engine = rls.Instance.Params

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type PolicyEngineFull struct {
	InfoMeta                        `json:",inline"`
	types.PolicyEngineComponentInfo `json:"info"`
	// mq连接对象绑定的内建消息队列对象实例配置,详情见各类内建消息队列配置,类型见info字段中的MQType字段
	Instance json.RawMessage `json:"instance,omitempty"`
}

// GetPolicyEngineInfo 获取policyengine类型示例信息
//
//	@Summary	获取policyengine类型示例信息
//
//	@Tags		proton,connect,policyengine
//	@Accept		json
//	@Produce	json
//
//
//	@Param		name									path		string				true	"policyengine"
//
//	@Success	200										{object}	PolicyEngineFull	"policyengine"
//	@Failure	400										{object}	HTTPError			"客户端请求参数错误"
//	@Failure	404										{object}	HTTPError			"对象不存在"
//	@Failure	500										{object}	HTTPError			"系统内部错误"
//	@Router		/components/info/policyengine/{name}	[get]
func (s *Server) GetPolicyEngineInfo(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("release name must not be empty").abortGin(ctx)
	// 	return
	// }
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if conf.ResourceConnectInfo.PolicyEngine == nil {
		NotFoundError.AbortGin(ctx)
		return
	}
	info := conf.ResourceConnectInfo.PolicyEngine
	res := &PolicyEngineFull{
		InfoMeta: InfoMeta{
			Name: "policyengine",
		},
		PolicyEngineComponentInfo: *info,
	}
	if info.SourceType == "internal" {
		// TODO remove
		rls := &cmp.ComponentInstance[types.ComponentPolicyEngine]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: pleRlsName,
				Type: "policyengine",
			},
		}
		err := cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		ins := rls.Instance
		bs, rerr := json.Marshal(ins)
		if rerr != nil {
			UnknownError.From("decode policyengine param error: " + rerr.Error()).AbortGin(ctx)
			return
		}
		res.Instance = bs
	}
	ctx.JSON(http.StatusOK, res)
}

// UpdatePolicyEngineInfo update policyengine info
//
//	@Summary		更新policyengine连接信息
//	@Description	更新policyengine连接信息
//
//	@Tags			proton,connect,policyengine
//	@Accept			json
//	@Produce		json
//
//	@Param			obj								body		PolicyEngineFull	true	"policyengine"
//
//	@Success		200								{object}	int					"返回为空"
//	@Failure		400								{object}	HTTPError			"客户端请求参数错误"
//	@Failure		404								{object}	HTTPError			"对象不存在，不允许更新"
//	@Failure		500								{object}	HTTPError			"系统内部错误"
//	@Router			/components/info/policyengine 	[PUT]
func (s *Server) UpdatePolicyEngineInfo(ctx *gin.Context) {
	body := &PolicyEngineFull{}
	if rerr := ctx.BindJSON(body); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	body.Name = "policyengine"
	obj := body.PolicyEngineComponentInfo

	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if obj.SourceType == "internal" {
		kupdate := true
		rls := &cmp.ComponentInstance[types.ComponentPolicyEngine]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: pleRlsName,
				Type: "policyengine",
			},
		}
		if body.Instance == nil {
			err := cmp.Get(ctx, s.ccli, rls)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			kupdate = false
		} else {
			if rerr := json.Unmarshal(body.Instance, &rls.Instance); rerr != nil {
				ParamError.From("decode kafka instance error: " + rerr.Error()).AbortGin(ctx)
				return
			}
		}
		if rls.Name == "" {
			rls.Name = pleRlsName
		}
		if rls.Instance.Dependencies == nil || rls.Instance.Dependencies.ETCD == "" {
			rls.Instance.Dependencies = &components.PolicyEngineComponentDependencies{
				ETCD: etcdRlsName,
			}
		}

		// update kafka
		if kupdate {
			UpdateInfoBindReleaseMarco(ctx, s, rls)
			if ctx.IsAborted() {
				return
			}
		}

		err := cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		conf.ResourceConnectInfo.PolicyEngine = rls.Instance.Info
		conf.Proton_policy_engine = rls.Instance.Params
	} else {
		conf.ResourceConnectInfo.PolicyEngine = &obj
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}
