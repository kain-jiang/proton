package proton_component

import (
	"net/http"

	"component-manage/pkg/models/types"
	"taskrunner/api/rest/proton_component/cmp"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// GetProtonNebula get nebula
//
//	@Summary	获取proton的nebula实例
//
//	@Tags		proton,ProtonCompoment,nebula
//	@Accept		json
//	@Produce	json
//
//	@Param		name								path		string					true	"实例名称"
//
//	@Success	200									{object}	types.ComponentNebula	"实例数据"
//	@Failure	400									{object}	HTTPError				"客户端请求参数错误"
//	@Failure	404									{object}	HTTPError				"对象不存在"
//	@Failure	500									{object}	HTTPError				"系统内部错误"
//	@Router		/components/release/nebula/{name}	[get]
func (s *Server) GetProtonNebula(ctx *gin.Context) {
	// name := ctx.Param("name")
	// if name == "" {
	// 	paramError.From("must set zookeeper realse name").abortGin(ctx)
	// 	return
	// }
	name := nebulaRlsName

	rls := &cmp.ComponentInstance[types.ComponentNebula]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: name,
			Type: "nebula",
		},
	}
	GetProtonReleaseMarco(ctx, s, rls)
	if !ctx.IsAborted() {
		raw := ctx.DefaultQuery("raw", "")
		if raw == "" || raw == "false" {
			rls.Instance.Params.Password = ""
			rls.Instance.Info.Password = ""
		}
		ctx.JSON(http.StatusOK, rls.Instance)
	}
}

// UpdateNebula update proton nebula instance
//
//	@Summary		更新proton nebula
//	@Description	更新proton nebula
//
//	@Tags			proton,ProtonCompoment,nebula
//	@Accept			json
//	@Produce		json
//
//	@Param			obj							body		types.ComponentNebula	true	"zookeeper配置信息"
//
//	@Success		200							{object}	int						"返回为空"
//	@Failure		400							{object}	HTTPError				"客户端请求参数错误"
//	@Failure		404							{object}	HTTPError				"对象不存在，不允许更新"
//	@Failure		500							{object}	HTTPError				"系统内部错误"
//	@Router			/components/release/nebula 	[PUT]
func (s *Server) UpdateProtonNebula(ctx *gin.Context) {
	obj := &types.ComponentNebula{}
	if rerr := ctx.BindJSON(obj); rerr != nil {
		ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}

	// if obj.Name == "" {
	// 	paramError.From("name must be not empty").abortGin(ctx)
	// 	return
	// }
	obj.Name = nebulaRlsName

	rls := cmp.ComponentInstance[types.ComponentNebula]{
		ComponentInstanceMeta: cmp.ComponentInstanceMeta{
			Name: obj.Name,
			Type: "nebula",
		},
		// Instance: *obj,
	}
	if obj.Params.Password == "" {
		if err := cmp.Get(ctx, s.ccli, &rls); err != nil {
			if !trait.IsInternalError(err, trait.ErrNotFound) {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
		} else {
			obj.Params.Password = rls.Instance.Params.Password
		}
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

	conf.Nebula = rls.Instance.Params

	if err := s.pcli.SetFullConf(ctx, conf); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}
