package cerr

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// E 自定义错误
type E struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Cause   string `json:"cause"`
}

func NewError(code int, message string, cause string) E {
	return E{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func AsError(err error) E {
	var rel E
	if errors.As(err, &rel) {
		return rel
	}
	return NewError(
		ServerProduceError,
		"server produce internal error",
		err.Error(),
	)
}

// HCode Http状态码
func (e E) HCode() int {
	c, _ := strconv.Atoi(strconv.Itoa(e.Code)[:3])
	return c
}

// Error 实现 error 接口
func (e E) Error() string {
	return fmt.Sprintf("code=%d, message='%s', cause='%s'", e.Code, e.Message, e.Cause)
}

// Reply 响应错误
func (e E) Reply(c *gin.Context) {
	c.JSON(e.HCode(), e)
	c.Abort()
}

// Reply 响应
func Reply(err error, c *gin.Context) {
	AsError(err).Reply(c)
}
