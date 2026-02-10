package rest

import (
	"fmt"
	"net/http"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// ListWorkApplicationInstance list work application instance
//
//	@Summary		获取已安装应用实例列表
//	@Description	获取已安装应用实例列表，支持的过滤参数有name与status，分别过滤应用名称与状态。
//	@Description	当name参数非空时，以name为筛选条件。
//	@Description	当status非空时，以status为筛选条件。
//	@Description	当两个筛选条件为空时，为正常获取当前已安装应用实例列表
//
//	@Tags			applicationInstance
//	@Tags			application
//	@Accept			json
//	@Produce		json
//
//	@Param			offset						query		int				true	"分页偏移量"
//	@Param			sid							query		int				true	"系统ID,多实例模式必填"
//	@Param			name						query		string			false	"应用名称,区分大小写"
//	@Param			status						query		[]int			false	"状态过滤器设置,多个状态间关系为'或关系'"
//	@Param			limit						query		int				true	"分页大小"
//	@Param			lang						query		string			false	"语言参数"
//	@Param			title						query		string			false	"对应语言的包名"
//
//	@Success		200							{object}	pageWorkAppIns	"工作应用实例分页"
//	@Failure		400							{object}	HTTPError		"客户端请求参数错误"
//	@Failure		500							{object}	HTTPError		"系统内部错误"
//	@Router			/application/instance/work 	[get]
func (e *ExecutorEngine) ListWorkApplicationInstance(ctx *gin.Context) {
	name := ctx.Query("name")
	status, err1 := ConvertStringToIntArray(ctx.QueryArray("status")...)
	if err1 != nil {
		ParamError.From(fmt.Sprintf("status must set with int array, convert error: %s", err1.Error())).AbortGin(ctx)
		return
	}
	sid := 0
	if e.SID >= 0 {
		sid = e.SID
	} else {
		queryInt, err := ParseIntFromQueryWithDefault(ctx, []string{"sid"}, "-1")
		if err != nil {
			err.AbortGin(ctx)
			return
		}
		sid = queryInt[0]
	}

	query, err := parseIntFromQuery(ctx, "limit", "offset")
	if err != nil {
		err.AbortGin(ctx)
		return
	}

	if name == "" {
		lang := ctx.Query("lang")
		alias := ctx.Query("title")
		aname := e.Store.GetAname(lang, alias, AppZone)
		if aname == "" {
			aname = alias
		}
		name = aname
	}

	filter := &trait.AppInsFilter{
		Status: status,
		Name:   name,
		Sid:    sid,
		Limit:  query[0],
		Offset: query[1],
	}
	am, err0 := e.ListWorkAPPIns(ctx, filter)
	if err0 != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}
	filter.Offset = 0
	total, err0 := e.CountWorkAppIns(ctx, filter)
	if err0 != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}
	if lang := ctx.Query("lang"); lang != "" {
		for i, j := range am {
			alias, err := e.GetAppLang(ctx, lang, j.AName, AppZone)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			am[i].Alias = alias
		}
	}
	ctx.JSON(http.StatusOK, pageWorkAppIns{
		TotalNum: total,
		Data:     am,
	})
}

type pageWorkAppIns struct {
	// 当前状态与查询条件下总共搜索到的数据条数
	TotalNum int `json:"totalNum"`
	// 当前分页已工作任务概要信息列表
	Data []trait.ApplicationInstanceOverview `json:"data"`
}
