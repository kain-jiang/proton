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

// GetProtonMariaDB get rds
//
//	@Summary	获取proton的rds实例
//
//	@Tags		proton,ProtonCompoment,mariadb
//	@Accept		json
//	@Produce	json
//
//	@Param		name								path		string					true	"实例名称"
//
//	@Success	200									{object}	types.ComponentMariaDB	"实例数据"
//	@Failure	400									{object}	HTTPError				"客户端请求参数错误"
//	@Failure	404									{object}	HTTPError				"对象不存在"
//	@Failure	500									{object}	HTTPError				"系统内部错误"
//	@Router		/components/release/mariadb/{name}	[get]
func (s *Server) GetProtonMariaDB(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("must set zookeeper realse name").abortGin(ctx)
	// 	return
	// }
	name := mariadbRlsName

	rls := &cmp.ComponentInstance[types.ComponentMariaDB]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "mariadb",
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

// UpdateMariaDB update proton rds instance
//
//	@Summary		更新proton rds
//	@Description	更新proton rds
//
//	@Tags			proton,ProtonCompoment,mariadb
//	@Accept			json
//	@Produce		json
//
//	@Param			obj								body		types.ComponentMariaDB	true	"zookeeper配置信息"
//
//	@Success		200								{object}	int						"返回为空"
//	@Failure		400								{object}	HTTPError				"客户端请求参数错误"
//	@Failure		404								{object}	HTTPError				"对象不存在，不允许更新"
//	@Failure		500								{object}	HTTPError				"系统内部错误"
//	@Router			/components/release/mariadb 	[PUT]
func (s *Server) UpdateProtonMariaDB(ctx *gin.Context) {
	obj := &types.ComponentMariaDB{}
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
	obj.Name = mariadbRlsName

	rls := cmp.ComponentInstance[types.ComponentMariaDB]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "mariadb",
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

	conf.Proton_mariadb = rls.Instance.Params
	if conf.ResourceConnectInfo.Rds != nil {
		if conf.ResourceConnectInfo.Rds.SourceType == "internal" {
			if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			conf.ResourceConnectInfo.Rds = rls.Instance.Info
		}
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

type RdsFull struct {
	InfoMeta                   `json:",inline"`
	types.MariaDBComponentInfo `json:"info"`
	// mq连接对象绑定的内建消息队列对象实例配置,详情见各类内建消息队列配置,类型见info字段中的MQType字段
	Instance *types.ComponentMariaDB `json:"instance,omitempty"`
}

// GetRdsInfo 获取rds类型示例信息
//
//	@Summary	获取rds类型示例信息
//
//	@Tags		proton,connect,rds
//	@Accept		json
//	@Produce	json
//
//
//	@Param		name						path		string		true	"rds"
//
//	@Success	200							{object}	RdsFull		"rds"
//	@Failure	400							{object}	HTTPError	"客户端请求参数错误"
//	@Failure	404							{object}	HTTPError	"对象不存在"
//	@Failure	500							{object}	HTTPError	"系统内部错误"
//	@Router		/components/info/rds/{name}	[get]
func (s *Server) GetRdsInfo(ctx *gin.Context) {
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

	if conf.ResourceConnectInfo.Rds == nil {
		NotFoundError.AbortGin(ctx)
		return
	}
	info := conf.ResourceConnectInfo.Rds
	res := &RdsFull{
		InfoMeta: InfoMeta{
			Name: "rds",
		},
		MariaDBComponentInfo: *info,
	}
	if info.SourceType == "internal" {
		// TODO remove
		rls := &cmp.ComponentInstance[types.ComponentMariaDB]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: mariadbRlsName,
				Type: "mariadb",
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
		res.MariaDBComponentInfo.Password = ""
	}
	ctx.JSON(http.StatusOK, res)
}

// UpdateRdsInfo update rds info
//
//	@Summary		更新rds连接信息
//	@Description	更新rds连接信息
//
//	@Tags			proton,connect,rds
//	@Accept			json
//	@Produce		json
//
//	@Param			obj						body		RdsFull		true	"rds"
//
//	@Success		200						{object}	int			"返回为空"
//	@Failure		400						{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		500						{object}	HTTPError	"系统内部错误"
//	@Router			/components/info/rds 	[PUT]
func (s *Server) UpdateRdsInfo(ctx *gin.Context) {
	body := &RdsFull{}
	bs, rerr := io.ReadAll(ctx.Request.Body)
	if rerr != nil {
		UnknownError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	if rerr := json.Unmarshal(bs, body); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	body.Name = "rds"
	obj := body.MariaDBComponentInfo

	// store into proton cli config
	conf, err := s.pcli.GetFullConf(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if obj.SourceType == "internal" {
		kupdate := true
		rls := &cmp.ComponentInstance[types.ComponentMariaDB]{
			ComponentInstanceMeta: cmp.ComponentInstanceMeta{
				Name: mariadbRlsName,
				Type: "mariadb",
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
			rls.Name = mariadbRlsName
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
		conf.ResourceConnectInfo.Rds = rls.Instance.Info
		conf.Proton_mariadb = rls.Instance.Params
	} else {
		if obj.Password == "" && conf.ResourceConnectInfo.Rds != nil {
			obj.Password = conf.ResourceConnectInfo.Rds.Password
		}
		conf.ResourceConnectInfo.Rds = &obj
	}

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
}
