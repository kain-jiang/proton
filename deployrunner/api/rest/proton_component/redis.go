package proton_component

import (
	"net/http"

	"component-manage/pkg/models/types"
	"taskrunner/api/rest/proton_component/cmp"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// GetProtonRedis get redis
//
//	@Summary	获取proton的redis实例
//
//	@Tags		proton,ProtonCompoment,redis
//	@Accept		json
//	@Produce	json
//
//	@Param		name								path		string					true	"实例名称"
//
//	@Success	200									{object}	types.ComponentRedis	"实例数据"
//	@Failure	400									{object}	HTTPError				"客户端请求参数错误"
//	@Failure	404									{object}	HTTPError				"对象不存在"
//	@Failure	500									{object}	HTTPError				"系统内部错误"
//	@Router		/components/release/redis/{name}	[get]
func (s *Server) GetProtonRedis(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("must set zookeeper realse name").abortGin(ctx)
	// 	return
	// }
	name := redisRlsName

	rls := &cmp.ComponentInstance[types.ComponentRedis]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "redis",
		},
	}
	GetProtonReleaseMarco(ctx, s, rls)
	if !ctx.IsAborted() {
		raw := ctx.DefaultQuery("raw", "")
		if raw == "" || raw == "false" {
			rls.Instance.Params.Admin_passwd = ""
			rls.Instance.Info.Password = ""
			rls.Instance.Info.SentinelPassword = ""
		}
		ctx.JSON(http.StatusOK, rls.Instance)
	}
}

// UpdateRedis update proton redis instance
//
//	@Summary		更新proton redis
//	@Description	更新proton redis
//
//	@Tags			proton,ProtonCompoment,redis
//	@Accept			json
//	@Produce		json
//
//	@Param			obj							body		types.ComponentRedis	true	"zookeeper配置信息"
//
//	@Success		200							{object}	int						"返回为空"
//	@Failure		400							{object}	HTTPError				"客户端请求参数错误"
//	@Failure		404							{object}	HTTPError				"对象不存在，不允许更新"
//	@Failure		500							{object}	HTTPError				"系统内部错误"
//	@Router			/components/release/redis 	[PUT]
func (s *Server) UpdateProtonRedis(ctx *gin.Context) {
	obj := &types.ComponentRedis{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	// if obj.Name == "" {
	// 	paramError.From("name must be not empty").abortGin(ctx)
	// 	return
	// }
	obj.Name = redisRlsName

	rls := cmp.ComponentInstance[types.ComponentRedis]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "redis",
		},
		// Instance: *obj,
	}
	if obj.Params.Admin_passwd == "" {
		if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
			if !trait.IsInternalError(err, trait.ErrNotFound) {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
		}
		obj.Params.Admin_passwd = rls.Instance.Params.Admin_passwd
	}
	rls.Instance = *obj
	UpdateProtonReleaseMarco(ctx, s, rls)
	if ctx.IsAborted() {
		return
	}

	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	conf.Proton_redis = rls.Instance.Params
	if conf.ResourceConnectInfo.Redis != nil {
		if conf.ResourceConnectInfo.Redis.SourceType == "internal" {
			if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			conf.ResourceConnectInfo.Redis = rls.Instance.Info
		}
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type RedisFull struct {
	InfoMeta                 `json:",inline"`
	types.RedisComponentInfo `json:"info"`
	// mq连接对象绑定的内建消息队列对象实例配置,详情见各类内建消息队列配置,类型见info字段中的MQType字段
	Instance *types.ComponentRedis `json:"instance,omitempty"`
}

// GetRedisInfo 获取redis类型示例信息
//
//	@Summary	获取redis类型示例信息
//
//	@Tags		proton,connect,redis
//	@Accept		json
//	@Produce	json
//
//
//	@Param		name							path		string		true	"redis"
//
//	@Success	200								{object}	RedisFull	"redis"
//	@Failure	400								{object}	HTTPError	"客户端请求参数错误"
//	@Failure	404								{object}	HTTPError	"对象不存在"
//	@Failure	500								{object}	HTTPError	"系统内部错误"
//	@Router		/components/info/redis/{name}	[get]
func (s *Server) GetRedisInfo(ctx *gin.Context) {
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

	if conf.ResourceConnectInfo.Redis == nil {
		NotFoundError.AbortGin(ctx)
		return
	}
	info := conf.ResourceConnectInfo.Redis
	res := &RedisFull{
		InfoMeta: InfoMeta{
			Name: "redis",
		},
		RedisComponentInfo: *info,
	}
	if info.SourceType == "internal" {
		// TODO remove
		rls := &cmp.ComponentInstance[types.ComponentRedis]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: redisRlsName,
				Type: "redis",
			},
		}
		err := cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		if raw == "" || raw == "false" {
			rls.Instance.Params.Admin_passwd = ""
			rls.Instance.Info.Password = ""
			rls.Instance.Info.SentinelPassword = ""
		}
		res.Instance = &rls.Instance
	}
	if raw == "" || raw == "false" {
		res.RedisComponentInfo.Password = ""
		res.RedisComponentInfo.SentinelPassword = ""
	}
	ctx.JSON(http.StatusOK, res)
}

// UpdateRedisInfo update redis info
//
//	@Summary		更新redis连接信息
//	@Description	更新redis连接信息
//
//	@Tags			proton,connect,redis
//	@Accept			json
//	@Produce		json
//
//	@Param			obj						body		RedisFull	true	"redis"
//
//	@Success		200						{object}	int			"返回为空"
//	@Failure		400						{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		500						{object}	HTTPError	"系统内部错误"
//	@Router			/components/info/redis 	[PUT]
func (s *Server) UpdateRedisInfo(ctx *gin.Context) {
	body := &RedisFull{}
	if rerr := ctx.BindJSON(body); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	body.Name = "redis"
	obj := body.RedisComponentInfo

	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if obj.SourceType == "internal" {
		kupdate := true
		rls := &cmp.ComponentInstance[types.ComponentRedis]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: redisRlsName,
				Type: "redis",
			},
		}
		err := cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			if !trait.IsInternalError(err, trait.ErrNotFound) {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
		}
		if body.Instance == nil {
			if rls.Instance.Params == nil {
				ParamError.From("instance param must set").AbortGin(ctx)
				return
			}
			kupdate = false
		} else {
			if body.Instance.Params.Admin_passwd == "" {
				body.Instance.Params.Admin_passwd = rls.Instance.Params.Admin_passwd
			}
			rls.Instance = *body.Instance
		}
		if rls.Name == "" {
			rls.Name = "redis"
		}

		// update kafka
		if kupdate {
			UpdateInfoBindReleaseMarco(ctx, s, rls)
			if ctx.IsAborted() {
				return
			}
		}

		err = cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		conf.ResourceConnectInfo.Redis = rls.Instance.Info
		conf.Proton_redis = rls.Instance.Params
	} else {
		if (obj.Password == "" || obj.SentinelPassword == "") && conf.ResourceConnectInfo.Redis != nil {
			obj.Password = conf.ResourceConnectInfo.Redis.SentinelPassword
			obj.SentinelPassword = conf.ResourceConnectInfo.Redis.SentinelPassword
		}
		conf.ResourceConnectInfo.Redis = &obj
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}
