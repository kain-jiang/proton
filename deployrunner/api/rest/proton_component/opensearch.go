package proton_component

import (
	"net/http"

	"component-manage/pkg/models/types"
	"taskrunner/api/rest/proton_component/cmp"

	"github.com/gin-gonic/gin"
)

// GetProtonOpensearch get opensearch
//
//	@Summary	获取proton的opensearch实例
//
//	@Tags		proton,ProtonCompoment,opensearch
//	@Accept		json
//	@Produce	json
//
//	@Param		name									path		string						true	"实例名称"
//
//	@Success	200										{object}	types.ComponentOpensearch	"实例数据"
//	@Failure	400										{object}	HTTPError					"客户端请求参数错误"
//	@Failure	404										{object}	HTTPError					"对象不存在"
//	@Failure	500										{object}	HTTPError					"系统内部错误"
//	@Router		/components/release/opensearch/{name}	[get]
func (s *Server) GetProtonOpensearch(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("must set zookeeper realse name").abortGin(ctx)
	// 	return
	// }
	name := opensearRlsName

	rls := &cmp.ComponentInstance[types.ComponentOpensearch]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "opensearch",
		},
	}
	GetProtonReleaseMarco(ctx, s, rls)
	if !ctx.IsAborted() {
		raw := ctx.DefaultQuery("raw", "")
		if raw == "" || raw == "false" {
			rls.Instance.Info.Password = ""
		}
		ctx.JSON(http.StatusOK, rls.Instance)
	}
}

// UpdateOpensearch update proton opensearch instance
//
//	@Summary		更新proton opensearch
//	@Description	更新proton opensearch
//
//	@Tags			proton,ProtonCompoment,opensearch
//	@Accept			json
//	@Produce		json
//
//	@Param			obj								body		types.ComponentOpensearch	true	"zookeeper配置信息"
//
//	@Success		200								{object}	int							"返回为空"
//	@Failure		400								{object}	HTTPError					"客户端请求参数错误"
//	@Failure		404								{object}	HTTPError					"对象不存在，不允许更新"
//	@Failure		500								{object}	HTTPError					"系统内部错误"
//	@Router			/components/release/opensearch 	[PUT]
func (s *Server) UpdateProtonOpensearch(ctx *gin.Context) {
	obj := &types.ComponentOpensearch{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	// if obj.Name == "" {
	// 	paramError.From("name must be not empty").abortGin(ctx)
	// 	return
	// }
	obj.Name = opensearRlsName

	rls := cmp.ComponentInstance[types.ComponentOpensearch]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "opensearch",
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

	conf.OpenSearch = rls.Instance.Params
	if conf.ResourceConnectInfo.OpenSearch != nil {
		if conf.ResourceConnectInfo.OpenSearch.SourceType == "internal" {
			if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			conf.ResourceConnectInfo.OpenSearch = rls.Instance.Info
		}
	}
	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type OpensearchFull struct {
	InfoMeta                      `json:",inline"`
	types.OpensearchComponentInfo `json:"info"`
	// mq连接对象绑定的内建消息队列对象实例配置,详情见各类内建消息队列配置,类型见info字段中的MQType字段
	Instance *types.ComponentOpensearch `json:"instance,omitempty"`
}

// GetOpensearchInfo 获取opensearch类型示例信息
//
//	@Summary	获取opensearch类型示例信息
//
//	@Tags		proton,connect,opensearch
//	@Accept		json
//	@Produce	json
//
//
//	@Param		name								path		string			true	"opensearch-master"
//
//	@Success	200									{object}	OpensearchFull	"opensearch"
//	@Failure	400									{object}	HTTPError		"客户端请求参数错误"
//	@Failure	404									{object}	HTTPError		"对象不存在"
//	@Failure	500									{object}	HTTPError		"系统内部错误"
//	@Router		/components/info/opensearch/{name}	[get]
func (s *Server) GetOpensearchInfo(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("release name must not be empty").abortGin(ctx)
	// 	return
	// }
	raw := ctx.DefaultQuery("raw", "")

	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if conf.ResourceConnectInfo.OpenSearch == nil {
		NotFoundError.AbortGin(ctx)
		return
	}
	info := conf.ResourceConnectInfo.OpenSearch
	res := &OpensearchFull{
		InfoMeta: InfoMeta{
			Name: "opensearch",
		},
		OpensearchComponentInfo: *info,
	}
	if info.SourceType == "internal" {
		// TODO remove
		rls := &cmp.ComponentInstance[types.ComponentOpensearch]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: opensearRlsName,
				Type: "opensearch",
			},
		}
		err := cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		if raw == "" || raw == "false" {
			rls.Instance.Info.Password = ""
		}
		res.Instance = &rls.Instance
	}
	if raw == "" || raw == "false" {
		res.OpensearchComponentInfo.Password = ""
	}
	ctx.JSON(http.StatusOK, res)
}

// UpdateOpensearchInfo update opensearch info
//
//	@Summary		更新opensearch连接信息
//	@Description	更新opensearch连接信息
//
//	@Tags			proton,connect,opensearch
//	@Accept			json
//	@Produce		json
//
//	@Param			obj								body		OpensearchFull	true	"opensearch"
//
//	@Success		200								{object}	int				"返回为空"
//	@Failure		400								{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404								{object}	HTTPError		"对象不存在，不允许更新"
//	@Failure		500								{object}	HTTPError		"系统内部错误"
//	@Router			/components/info/opensearch 	[PUT]
func (s *Server) UpdateOpensearchInfo(ctx *gin.Context) {
	body := &OpensearchFull{}
	if rerr := ctx.BindJSON(body); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	body.Name = "opensearch"
	obj := body.OpensearchComponentInfo

	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if obj.SourceType == "internal" {
		kupdate := true
		rls := &cmp.ComponentInstance[types.ComponentOpensearch]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: opensearRlsName,
				Type: "opensearch",
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
			rls.Instance = *body.Instance
		}
		if rls.Name == "" {
			rls.Name = "opensearch"
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
		conf.ResourceConnectInfo.OpenSearch = rls.Instance.Info
		conf.OpenSearch = rls.Instance.Params
	} else {
		if obj.Password == "" && conf.ResourceConnectInfo.OpenSearch != nil {
			obj.Password = conf.ResourceConnectInfo.OpenSearch.Password
		}
		conf.ResourceConnectInfo.OpenSearch = &obj
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}
