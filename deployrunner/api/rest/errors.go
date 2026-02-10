package rest

import (
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

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
		ErrorCode:  e.ErrorCode,
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

	ComponentNotfoundError = HTTPError{
		StatusCode: 412,
		Detail:     "component not found",
		ErrorCode:  trait.ErrComponentNotFound,
	}

	UnknownError = HTTPError{
		StatusCode: 500,
		ErrorCode:  500,
		Detail:     "internal error",
	}

	UniqueKeyError = HTTPError{
		StatusCode: 409,
		ErrorCode:  409,
		Detail:     "ths resouce conflict with other ",
	}

	ConditionError = HTTPError{
		StatusCode: 412,
		ErrorCode:  412,
		Detail:     "the condition for operation isn't ok",
	}

	ApplicationStillUseError = HTTPError{
		StatusCode: 412,
		ErrorCode:  trait.ErrApplicationStillUse,
		Detail:     "the condition for operation isn't ok",
	}

	ClientTimeoutError = HTTPError{
		StatusCode: 408,
		ErrorCode:  408,
		Detail:     "the client request arrival timeout",
	}

	engineProxyConnectError = HTTPError{
		StatusCode: 503,
		ErrorCode:  503,
		Detail:     "the engine request connect error",
	}

	engineProxyTimeoutError = HTTPError{
		StatusCode: 504,
		ErrorCode:  504,
		Detail:     "the engine proxy request not reponse before request cancel",
	}

	// AuthenticationError 无权限，但由于原有项目的错误使用，使用401为错误码
	AuthenticationError = HTTPError{
		StatusCode: 401,
		ErrorCode:  401,
		Detail:     "the user can't operate without permission",
	}

	// AuthenticationError 认证失败，但由于原有项目的错误使用，使用403为错误码
	InvalidAuthorized = HTTPError{
		StatusCode: 403,
		ErrorCode:  403,
		Detail:     "invalid authorize",
	}

	ApplicationNotfoundError = HTTPError{
		StatusCode: 412,
		Detail:     "component not found",
		ErrorCode:  trait.ErrApplicationNotFound,
	}
)
