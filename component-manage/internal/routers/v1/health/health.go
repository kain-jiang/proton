package health

import (
	"net/http"

	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiHealthAlive 存活检查
//
//	@Summary		存活检查
//	@Description	健康检查接口，存活检查
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	response.Status
//	@Router			/health/alive [get]
func ApiHealthAlive(c *gin.Context) {
	c.JSON(http.StatusOK, response.OK)
}

// ApiHealthReady 就绪检查
//
//	@Summary		就绪检查
//	@Description	健康检查接口，就绪检查
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	response.Status
//	@Router			/health/ready [get]
func ApiHealthReady(c *gin.Context) {
	c.JSON(http.StatusOK, response.OK)
}
