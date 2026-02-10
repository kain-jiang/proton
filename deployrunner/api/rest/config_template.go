package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// UploadConfigTemplate the application's config template
//
//	@Summary		上传一个应用配置模板
//	@Description	上传一个应用配置模板，该模板可在以当前系统运行应用实例配置时，与之合并返回。
//	@Description	一个应用配置模板以<aname, tname, tversion>作为唯一索引，如果该模板已存在，则覆盖已存在模板。
//	@Description	覆盖逻辑中，组合唯一索引字段外将会被更新。
//
//	@Tags			application
//	@Accept			json
//	@Produce		json
//
//	@Param			config					body		trait.AppliacationConfigTemplate	true	"应用配置模板内容"
//
//	@Success		200						{object}	int									"配置模板ID"
//	@Failure		500						{object}	HTTPError							"系统内部错误"
//	@Router			/application/config 	[post]
func (e *ExecutorEngine) UploadConfigTemplate(ctx *gin.Context) {
	meta := &trait.AppliacationConfigTemplate{}
	if err0 := ctx.BindJSON(meta); err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}

	if fe := CheckStringFieldIsEmpty(
		[]string{"aname", "aversion", "tname", "tversion"},
		meta.Aname, meta.Aversion, meta.Tname, meta.Tversion); fe != "" {
		ParamError.From(fe + ", the field must set").AbortGin(ctx)
	}

	id, err := e.Store.InsertConfigTempalte(ctx, *meta)
	if trait.IsInternalError(err, trait.ErrParam) {
		ParamError.From(err.Error()).AbortGin(ctx)
		return
	}
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, id)
}

// GetConfigTemplate get the application config template
//
//	@Summary		获取应用配置模板详细内容
//	@Description	获取应用配置模板详细内容
//
//	@Tags			application
//	@Accept			json
//	@Produce		json
//
//	@Param			tid							path		int									true	"应用配置模板ID"
//
//	@Success		200							{object}	trait.AppliacationConfigTemplate	"配置模板详细信息"
//	@Failure		400							{object}	HTTPError							"客户端请求参数错误"
//	@Failure		404							{object}	HTTPError							"对象不存在"
//	@Failure		500							{object}	HTTPError							"系统内部错误"
//	@Router			/application/config/{tid} 	[get]
func (e *ExecutorEngine) GetConfigTemplate(ctx *gin.Context) {
	idStr := ctx.Param("tid")
	id, err0 := strconv.Atoi(idStr)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the tid is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}

	obj, err := e.Store.GetConfigTemplate(ctx, id)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, obj)
}

// ListConfigTemplate list config template
//
//	@Summary		根据过滤需要获取配置模板
//	@Description	该接口主要用于从任务视图或组件视图视角查看任务执行信息
//
//	@Tags			application
//	@Accept			json
//	@Produce		json
//
//	@Param			offset				query		int					false	"分页偏移量"
//	@Param			limit				query		int					false	"分页大小"
//	@Param			l					query		[]string			false	"设置标签过滤器中标签,如'http://127.0.0.1/test?l=test0&l=test1'"
//	@Param			lc					query		int					false	"标签过滤器类型，0或1或不设置时为'或'关系，2为'与关系'"
//	@Param			v					query		string				false	"设置版本过滤器中版本，非空时启用"
//	@Param			vt					query		int					false	"版本过滤器类型,0或1或不设置时为精准匹配,2为patch版本匹配,3为大于等于当前版本匹配"
//	@Param			aname				query		string				false	"应用名过滤条件,为空时其余过滤器不可设置"
//	@Param			count				query		string				false	"是否返回对应过滤条件下的数据数量，false或False为否，其余为是，一般在分页查询仅第一次查询时设置为是"
//
//	@Success		200					{object}	ConfigTempalteList	"结果列表"
//	@Failure		400					{object}	HTTPError			"客户端请求参数错误"
//	@Failure		500					{object}	HTTPError			"系统内部错误"
//	@Router			/application/config	[get]
func (e *ExecutorEngine) ListConfigTemplate(ctx *gin.Context) {
	needCount := ctx.DefaultQuery("count", "True")
	params, err0 := ParseIntFromQueryWithDefault(ctx,
		[]string{
			"offset",
			"limit",
			"lc",
			"vt",
		},
		"0",
		"10",
		"1",
		"1",
	)
	if err0 != nil {
		err0.AbortGin(ctx)
		return
	}
	labels := ctx.QueryArray("l")
	aname := ctx.Query("aname")
	aversion := ctx.Query("v")

	f := trait.ApplicationConfigTemplateFilter{
		Aname: aname,
	}
	if aversion != "" {
		f.ApplicationVersionFilter = &trait.ApplicationVersionFilter{
			Aversion: aversion,
			Type:     params[3],
		}
	}
	if len(labels) != 0 {
		f.ApplicationLabelFilter = &trait.ApplicationLabelFilter{
			Labels:    labels,
			Condition: params[2],
		}
	}

	res := &ConfigTempalteList{}
	if needCount != "False" && needCount != "false" {
		count, err1 := e.CountConfigTempalte(ctx, f)
		if trait.IsInternalError(err1, trait.ErrParam) {
			ParamError.From(err1.Error()).AbortGin(ctx)
			return
		}
		if err1 != nil {
			UnknownError.From(err1.Error()).AbortGin(ctx)
			return
		}
		res.TotalNum = count
	}

	objs, err1 := e.Store.ListConfigTemplate(ctx, f, params[1], params[0])
	if trait.IsInternalError(err1, trait.ErrParam) {
		ParamError.From(err1.Error()).AbortGin(ctx)
		return
	}
	if err1 != nil {
		UnknownError.From(err1.Error()).AbortGin(ctx)
		return
	}
	res.Data = objs
	ctx.JSON(http.StatusOK, res)
}

type ConfigTempalteList struct {
	TotalNum int                                    `json:"totalNum"`
	Data     []trait.AppliacationConfigTemplateMeta `json:"data"`
}

// DeleteConfigTemplate delete the application config template
//
//	@Summary		删除应用配置模板
//	@Description	删除应用配置模板
//
//	@Tags			application
//	@Accept			json
//	@Produce		json
//
//	@Param			tid							path		int			true	"应用配置模板ID"
//
//	@Success		200							{object}	nil			"操作成功"
//	@Failure		400							{object}	HTTPError	"客户端请求参数错误"
//	@Failure		500							{object}	HTTPError	"系统内部错误"
//	@Router			/application/config/{tid} 	[delete]
func (e *ExecutorEngine) DeleteConfigTemplate(ctx *gin.Context) {
	idStr := ctx.Param("tid")
	id, err0 := strconv.Atoi(idStr)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the tid is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}

	err := e.Store.DeleteConfigTemplate(ctx, id)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, "")
}
