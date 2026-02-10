package rest

import (
	"fmt"
	"net/http"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// ListComponentInstanceDependence list depend component instance
//
//	@Summary		获取组件依赖的组件列表
//	@Description	获取对应组件依赖的组件列表
//
//	@Tags			componentInstance
//	@Accept			json
//	@Produce		json
//
//	@Param			cid										path		string						true	"应用实例ID"
//
//	@Success		200										{object}	[]trait.ComponentInstance	"依赖的组件实例对象列表"
//	@Failure		400										{object}	HTTPError					"客户端请求参数错误"
//	@Failure		500										{object}	HTTPError					"系统内部错误"
//	@Router			/component/instance/{cid}/dependence 	[get]
func (e *ExecutorEngine) ListComponentInstanceDependence(ctx *gin.Context) {
	cid, err0 := ConvertStringToIntArray(ctx.Param("cid"))
	if err0 != nil {
		ParamError.From(fmt.Sprintf("cid must set with int in path, convert error: %s", err0.Error())).AbortGin(ctx)
		return
	}

	cids, err := e.GetPointFrom(ctx, cid[0])
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	cs := make([]*trait.ComponentInstance, 0, len(cids))
	for _, id := range cids {
		c, err := e.GetComponentIns(ctx, id)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		cs = append(cs, c)

	}
	ctx.JSON(http.StatusOK, cs)
}
