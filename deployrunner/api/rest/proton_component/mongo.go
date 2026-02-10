package proton_component

import (
	"encoding/json"
	"io"
	"net/http"

	"component-manage/pkg/models/types"
	"taskrunner/api/rest/proton_component/cmp"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// GetProtonMongoDB get mongodb
//
//	@Summary	获取proton的mongodb实例
//
//	@Tags		proton,ProtonCompoment,mongodb
//	@Accept		json
//	@Produce	json
//
//	@Param		name								path		string					true	"实例名称"
//
//	@Success	200									{object}	types.ComponentMongoDB	"实例数据"
//	@Failure	400									{object}	HTTPError				"客户端请求参数错误"
//	@Failure	404									{object}	HTTPError				"对象不存在"
//	@Failure	500									{object}	HTTPError				"系统内部错误"
//	@Router		/components/release/mongodb/{name}	[get]
func (s *Server) GetProtonMongoDB(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("must set zookeeper realse name").abortGin(ctx)
	// 	return
	// }
	name := mongdbRlsName

	rls := &cmp.ComponentInstance[types.ComponentMongoDB]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "mongodb",
		},
	}
	GetProtonReleaseMarco(ctx, s, rls)
	if !ctx.IsAborted() {
		raw := ctx.DefaultQuery("raw", "")
		if raw == "" || raw == "false" {
			rls.Instance.Params.Admin_passwd = ""
			rls.Instance.Params.Password = ""
			rls.Instance.Info.Password = ""
		}
		ctx.JSON(http.StatusOK, rls.Instance)
	}
}

// UpdateMongoDB update proton mongodb instance
//
//	@Summary		更新proton mongodb
//	@Description	更新proton mongodb
//
//	@Tags			proton,ProtonCompoment,mongodb
//	@Accept			json
//	@Produce		json
//
//	@Param			obj								body		types.ComponentMongoDB	true	"zookeeper配置信息"
//
//	@Success		200								{object}	int						"返回为空"
//	@Failure		400								{object}	HTTPError				"客户端请求参数错误"
//	@Failure		404								{object}	HTTPError				"对象不存在，不允许更新"
//	@Failure		500								{object}	HTTPError				"系统内部错误"
//	@Router			/components/release/mongodb 	[PUT]
func (s *Server) UpdateProtonMongoDB(ctx *gin.Context) {
	obj := &types.ComponentMongoDB{}
	bs, rerr := io.ReadAll(ctx.Request.Body)
	if rerr != nil {
		UnknownError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	if rerr := json.Unmarshal(bs, obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	// if obj.Name == "" {
	// 	paramError.From("name must be not empty").abortGin(ctx)
	// 	return
	// }
	obj.Name = mongdbRlsName
	rls := cmp.ComponentInstance[types.ComponentMongoDB]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "mongodb",
		},
		// Instance: *obj,
	}
	if obj.Params.Password == "" || obj.Params.Admin_passwd == "" {
		if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
			if !trait.IsInternalError(err, trait.ErrNotFound) {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
		}
		obj.Params.Password = rls.Instance.Params.Password
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

	conf.Proton_mongodb = rls.Instance.Params
	if conf.ResourceConnectInfo.Mongodb != nil {
		if conf.ResourceConnectInfo.Mongodb.SourceType == "internal" {
			if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			conf.ResourceConnectInfo.Mongodb = rls.Instance.Info
		}
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type MongoDBFull struct {
	InfoMeta                   `json:",inline"`
	types.MongoDBComponentInfo `json:"info"`
	// mq连接对象绑定的内建消息队列对象实例配置,详情见各类内建消息队列配置,类型见info字段中的MQType字段
	Instance *types.ComponentMongoDB `json:"instance,omitempty"`
}

// GetMongoDBInfo 获取mongodb类型示例信息
//
//	@Summary	获取mongodb类型示例信息
//
//	@Tags		proton,connect,mongodb
//	@Accept		json
//	@Produce	json
//
//
//	@Param		name							path		string		true	"mongodb"
//
//	@Success	200								{object}	MongoDBFull	"mongodb"
//	@Failure	400								{object}	HTTPError	"客户端请求参数错误"
//	@Failure	404								{object}	HTTPError	"对象不存在"
//	@Failure	500								{object}	HTTPError	"系统内部错误"
//	@Router		/components/info/mongodb/{name}	[get]
func (s *Server) GetMongoDBInfo(ctx *gin.Context) {
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

	if conf.ResourceConnectInfo.Mongodb == nil {
		NotFoundError.AbortGin(ctx)
		return
	}
	info := conf.ResourceConnectInfo.Mongodb
	res := &MongoDBFull{
		InfoMeta: InfoMeta{
			Name: "mongodb",
		},
		MongoDBComponentInfo: *info,
	}
	if info.SourceType == "internal" {
		// TODO remove
		rls := &cmp.ComponentInstance[types.ComponentMongoDB]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: mongdbRlsName,
				Type: "mongodb",
			},
		}
		err := cmp.Get(ctx, s.ccli, rls)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		if raw == "" || raw == "false" {
			rls.Instance.Params.Admin_passwd = ""
			rls.Instance.Params.Password = ""
			rls.Instance.Info.Password = ""
		}
		res.Instance = &rls.Instance

	}
	if raw == "" || raw == "false" {
		res.MongoDBComponentInfo.Password = ""
	}
	ctx.JSON(http.StatusOK, res)
}

// UpdateMongoDBInfo update mongodb info
//
//	@Summary		更新mongodb连接信息
//	@Description	更新mongodb连接信息
//
//	@Tags			proton,connect,mongodb
//	@Accept			json
//	@Produce		json
//
//	@Param			obj							body		MongoDBFull	true	"mongodb"
//
//	@Success		200							{object}	int			"返回为空"
//	@Failure		400							{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404							{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		500							{object}	HTTPError	"系统内部错误"
//	@Router			/components/info/mongodb 	[PUT]
func (s *Server) UpdateMongoDBInfo(ctx *gin.Context) {
	body := &MongoDBFull{}
	bs, rerr := io.ReadAll(ctx.Request.Body)
	if rerr != nil {
		UnknownError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	if rerr := json.Unmarshal(bs, body); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	body.Name = "mongodb"
	obj := body.MongoDBComponentInfo

	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if obj.SourceType == "internal" {
		kupdate := true
		rls := &cmp.ComponentInstance[types.ComponentMongoDB]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: mongdbRlsName,
				Type: "mongodb",
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
			if body.Instance.Params.Admin_passwd == "" || body.Instance.Params.Password == "" {
				body.Instance.Params.Admin_passwd = rls.Instance.Params.Admin_passwd
				body.Instance.Params.Password = rls.Instance.Params.Password
			}
			rls.Instance = *body.Instance
		}
		if rls.Name == "" {
			rls.Name = mongdbRlsName
		}

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
		conf.ResourceConnectInfo.Mongodb = rls.Instance.Info
		conf.Proton_mongodb = rls.Instance.Params
	} else {
		if obj.Password == "" && conf.ResourceConnectInfo.Mongodb != nil {
			obj.Password = conf.ResourceConnectInfo.Mongodb.Password
		}
		conf.ResourceConnectInfo.Mongodb = &obj
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}
