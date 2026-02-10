package proton_component

import (
	"github.com/gin-gonic/gin"
)

// TODO MOVE httperror into common module
type HTTPError struct {
	// 状态码
	StatusCode int `json:"status"`
	// 错误分类下的错误码
	ErrorCode int `json:"code"`
	// 具体错误信息
	Detail string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Detail
}

func (e *HTTPError) From(detail string) *HTTPError {
	return &HTTPError{
		StatusCode: e.StatusCode,
		Detail:     detail,
	}
}

func (e *HTTPError) AbortGin(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(e.StatusCode, e)
}

var (
	ParamError = HTTPError{
		StatusCode: 400,
		ErrorCode:  400,
		Detail:     "param error, check the input",
	}

	NotFoundError = HTTPError{
		StatusCode: 404,
		ErrorCode:  404,
		Detail:     "the data not found",
	}

	UnknownError = HTTPError{
		StatusCode: 500,
		ErrorCode:  500,
		Detail:     "internal error",
	}
)
