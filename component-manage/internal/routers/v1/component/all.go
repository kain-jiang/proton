package component

import (
	"net/http"

	"component-manage/internal/logic"
	"component-manage/internal/pkg/cerr"

	"github.com/gin-gonic/gin"
)

// ApiComponentAllList 获取所有components
//
//	@Summary		获取所有components
//	@Description	获取所有components
//	@Schemes
//	@Tags		Component
//	@Produce	json
//	@Success	200	{object}	[]response.Component
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/all [get]
func ApiComponentAllList(c *gin.Context) {
	resp, err := logic.ListAllComponents()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
