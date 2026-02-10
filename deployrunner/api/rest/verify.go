package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// GetVerifyResult get the application's verify result
//
//	@Summary		获取验证记录
//	@Description	获取指定job的所有验证记录与结果
//
//	@Tags			verification
//	@Accept			json
//	@Produce		json
//
//	@Param			jid						path		int					true	"任务ID"
//
//	@Success		200						{object}	trait.VerifyRecord	"正常返回"
//	@Failure		400						{object}	HTTPError			"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError			"对象不存在"
//	@Failure		500						{object}	HTTPError			"系统内部错误"
//	@Router			/verification/{jid} 	[get]
func (e *ExecutorEngine) GetVerifyResult(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	if jidStr == "" {
		ParamError.From("jid is required").AbortGin(ctx)
		return
	}
	jid, err0 := strconv.Atoi(jidStr)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the offset is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}
	vb, err := e.Store.GetVerifyRecord(ctx, jid)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
	} else {
		ctx.JSON(http.StatusOK, vb)
	}
}

// GetFunctionTestEntries View Details of Single Data Validation
//
//	@Summary		查看单次功能验证详情
//	@Description	通过指定的数据验证fid，查看对应的测试用例
//
//	@Tags			verification
//	@Accept			json
//	@Produce		json
//
//	@Param			pageNum					query		int						false	"页数，默认1"
//	@Param			fid						query		string					true	"功能验证记录的标识id"
//	@Param			pageSize				query		int						false	"页面显示数量，默认10"
//
//	@Success		200						{object}	PageFunctionTestEntries	"正常返回"
//	@Failure		400						{object}	HTTPError				"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError				"对象不存在"
//	@Failure		500						{object}	HTTPError				"系统内部错误"
//	@Router			/verification/function 	[get]
func (e *ExecutorEngine) GetFunctionTestEntries(ctx *gin.Context) {
	fidStr := ctx.Query("fid")
	if fidStr == "" {
		ParamError.From("fid is required").AbortGin(ctx)
		return
	}

	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		ParamError.From(fmt.Sprintf("fid must set a integer: %s", err.Error())).AbortGin(ctx)
		return
	}

	query, err1 := parseIntFromQuery(ctx, "limit", "offset")
	if err1 != nil {
		err1.AbortGin(ctx)
		return
	}

	drs, err0 := e.Store.GetFunctionTestEntries(ctx, fid, query[0], query[1])
	if err0 != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}
	total, err0 := e.Store.CountFunctionTestEntries(ctx, fid)
	if err != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, PageFunctionTestEntries{
		TotalNum: total,
		Data:     drs,
	})
}

// GetDataTestEntries View Details of Single Data Validation
//
//	@Summary		查看单次数据验证详情
//	@Description	通过指定的数据验证did，查看对应的测试用例
//
//	@Tags			verification
//	@Accept			json
//	@Produce		json
//
//	@Param			pageNum					query		int					false	"页数,默认1"
//	@Param			did						query		string				true	"数据验证记录的标识id"
//	@Param			pageSize				query		int					false	"页面显示数量，默认10"
//
//	@Success		200						{object}	PageDataTestEntries	"正常返回"
//	@Failure		400						{object}	HTTPError			"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError			"对象不存在"
//	@Failure		500						{object}	HTTPError			"系统内部错误"
//	@Router			/verification/database 	[get]
func (e *ExecutorEngine) GetDataTestEntries(ctx *gin.Context) {
	didStr := ctx.Query("did")
	if didStr == "" {
		ParamError.From("did is required").AbortGin(ctx)
		return
	}

	did, err0 := strconv.Atoi(didStr)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("did must set a integer: %s", err0.Error())).AbortGin(ctx)
		return
	}

	query, err := parseIntFromQuery(ctx, "limit", "offset")
	if err != nil {
		err.AbortGin(ctx)
		return
	}

	drs, err1 := e.Store.GetDataTestEntries(ctx, did, query[0], query[1])
	if err1 != nil {
		UnknownError.From(err1.Error()).AbortGin(ctx)
		return
	}
	total, err1 := e.Store.CountDataTestEntries(ctx, did)
	if err1 != nil {
		UnknownError.From(err1.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, PageDataTestEntries{
		TotalNum: total,
		Data:     drs,
	})
}

// PageDataTestEntries 分页查询数据测试用例结果
type PageDataTestEntries struct {
	// 总共搜索到的数据条数
	TotalNum int `json:"totalNum"`
	// 当前页服务测试结果
	Data []trait.DataTestEntry `json:"data"`
}

// PageFunctionTestEntries 分页查询功能测试用例结果
type PageFunctionTestEntries struct {
	// 总共搜索到的数据条数
	TotalNum int `json:"totalNum"`
	// 当前页服务测试结果
	Data []trait.FunctionTestEntry `json:"data"`
}
