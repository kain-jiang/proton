package proton_component

import (
	"encoding/json"
	"net/http"

	"component-manage/pkg/models/types"
	"taskrunner/api/rest/proton_component/cmp"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// GetProtonEtcd get etcd
//
//	@Summary	获取proton的etcd实例
//
//	@Tags		proton,ProtonCompoment,etcd
//	@Accept		json
//	@Produce	json
//
//	@Param		name							path		string				true	"实例名称"
//
//	@Success	200								{object}	types.ComponentETCD	"实例数据"
//	@Failure	400								{object}	HTTPError			"客户端请求参数错误"
//	@Failure	404								{object}	HTTPError			"对象不存在"
//	@Failure	500								{object}	HTTPError			"系统内部错误"
//	@Router		/components/release/etcd/{name}	[get]
func (s *Server) GetProtonEtcd(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("must set zookeeper realse name").abortGin(ctx)
	// 	return
	// }
	name := etcdRlsName
	rls := &cmp.ComponentInstance[types.ComponentETCD]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "etcd",
		},
	}
	err := cmp.Get(ctx, s.ccli, rls)
	if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
		herr := HTTPError{
			StatusCode: err.Detail.(int),
			ErrorCode:  err.Detail.(int),
			Detail:     err.Error(),
		}
		herr.AbortGin(ctx)
		return
	}
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.AbortGin(ctx)
		return
	}
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, rls.Instance)
}

// UpdateEtcd update proton etcd instance
//
//	@Summary		更新proton etcd
//	@Description	更新proton etcd
//
//	@Tags			proton,ProtonCompoment,etcd
//	@Accept			json
//	@Produce		json
//
//	@Param			obj							body		types.ComponentETCD	true	"zookeeper配置信息"
//
//	@Success		200							{object}	int					"返回为空"
//	@Failure		400							{object}	HTTPError			"客户端请求参数错误"
//	@Failure		404							{object}	HTTPError			"对象不存在，不允许更新"
//	@Failure		500							{object}	HTTPError			"系统内部错误"
//	@Router			/components/release/etcd 	[PUT]
func (s *Server) UpdateProtonEtcd(ctx *gin.Context) {
	obj := &types.ComponentETCD{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	// if obj.Name == "" {
	// 	paramError.From("name must be not empty").abortGin(ctx)
	// 	return
	// }
	obj.Name = etcdRlsName

	rls := cmp.ComponentInstance[types.ComponentETCD]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "etcd",
		},
		Instance: *obj,
	}
	err := cmp.Update(ctx, s.ccli, &rls)
	if trait.IsInternalError(err, trait.ErrParam) {
		ParamError.From(err.Error()).AbortGin(ctx)
		return
	}
	if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
		herr := HTTPError{
			StatusCode: err.Detail.(int),
			ErrorCode:  err.Detail.(int),
			Detail:     "provider: component-manage, resp:" + err.Error(),
		}
		herr.AbortGin(ctx)
		return
	}
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	conf.Proton_etcd = rls.Instance.Params
	if conf.ResourceConnectInfo.Etcd != nil {
		if conf.ResourceConnectInfo.Etcd.SourceType == "internal" {
			if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			conf.ResourceConnectInfo.Etcd = rls.Instance.Info
		}
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type EtcdFull struct {
	InfoMeta                `json:",inline"`
	types.ETCDComponentInfo `json:"info"`
	// mq连接对象绑定的内建消息队列对象实例配置,详情见各类内建消息队列配置,类型见info字段中的MQType字段
	Instance json.RawMessage `json:"instance,omitempty"`
}

// GetEtcdInfo 获取etcd类型示例信息
//
//	@Summary	获取etcd类型示例信息
//
//	@Tags		proton,connect,etcd
//	@Accept		json
//	@Produce	json
//
//
//	@Param		name							path		string		true	"etcd"
//
//	@Success	200								{object}	EtcdFull	"etcd"
//	@Failure	400								{object}	HTTPError	"客户端请求参数错误"
//	@Failure	404								{object}	HTTPError	"对象不存在"
//	@Failure	500								{object}	HTTPError	"系统内部错误"
//	@Router		/components/info/etcd/{name}	[get]
func (s *Server) GetEtcdInfo(ctx *gin.Context) {
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

	if conf.ResourceConnectInfo.Etcd == nil {
		NotFoundError.AbortGin(ctx)
		return
	}
	info := conf.ResourceConnectInfo.Etcd
	res := &EtcdFull{
		InfoMeta: InfoMeta{
			Name: "etcd",
		},
		ETCDComponentInfo: *info,
	}
	if info.SourceType == "internal" {
		// TODO remove
		rls := &cmp.ComponentInstance[types.ComponentETCD]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: etcdRlsName,
				Type: "etcd",
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
			UnknownError.From("decode etcd param error: " + rerr.Error()).AbortGin(ctx)
			return
		}
		res.Instance = bs
	}
	ctx.JSON(http.StatusOK, res)
}

// UpdateEtcdInfo update etcd info
//
//	@Summary		更新etcd连接信息
//	@Description	更新etcd连接信息
//
//	@Tags			proton,connect,etcd
//	@Accept			json
//	@Produce		json
//
//	@Param			obj						body		EtcdFull	true	"etcd"
//
//	@Success		200						{object}	int			"返回为空"
//	@Failure		400						{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		500						{object}	HTTPError	"系统内部错误"
//	@Router			/components/info/etcd 	[PUT]
func (s *Server) UpdateEtcdInfo(ctx *gin.Context) {
	body := &EtcdFull{}
	if rerr := ctx.BindJSON(body); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	body.Name = "etcd"
	obj := body.ETCDComponentInfo

	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if obj.SourceType == "internal" {
		kupdate := true
		rls := &cmp.ComponentInstance[types.ComponentETCD]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: etcdRlsName,
				Type: "etcd",
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
			rls.Name = etcdRlsName
		}

		// update kafka
		if kupdate {
			if err := cmp.Update(ctx, s.ccli, rls); err != nil {
				if trait.IsInternalError(err, trait.ECHTTPAPIRawError) {
					herr := HTTPError{
						StatusCode: err.Detail.(int),
						ErrorCode:  err.Detail.(int),
						Detail:     "provider: component-manage, resp:" + err.Error(),
					}
					herr.AbortGin(ctx)
					return
				}
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
		}

		err := cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		conf.ResourceConnectInfo.Etcd = rls.Instance.Info
		conf.Proton_etcd = rls.Instance.Params
	} else {
		conf.ResourceConnectInfo.Etcd = &obj
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}
