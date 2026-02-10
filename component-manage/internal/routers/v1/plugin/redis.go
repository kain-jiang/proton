package plugin

import (
	"net/http"

	"component-manage/internal/logic/redis"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginRedisEnable 激活redis插件
//
//	@Summary		激活redis插件
//	@Description	激活redis插件，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Redis
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsRedis	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/redis [post]
func ApiPluginRedisEnable(c *gin.Context) {
	var req request.PluginsRedis
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := redis.EnableRedisPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginRedisUpgrade 更新redis插件信息
//
//	@Summary		更新redis插件
//	@Description	更新redis插件信息，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Redis
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsRedis	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/redis [put]
func ApiPluginRedisUpgrade(c *gin.Context) {
	var req request.PluginsRedis
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := redis.UpgradeRedisPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginRedisGet 获取redis插件信息
//
//	@Summary		获取redis插件
//	@Description	获取redis插件信息
//	@Schemes
//	@Tags		Plugin,Redis
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/redis [get]
func ApiPluginRedisGet(c *gin.Context) {
	resp, err := redis.GetRedisPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
