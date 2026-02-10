package plugin

import (
	"net/http"

	"component-manage/internal/logic/zookeeper"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginZookeeperEnable 激活zookeeper插件
//
//	@Summary		激活zookeeper插件
//	@Description	激活zookeeper插件，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Zookeeper
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsZookeeper	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/zookeeper [post]
func ApiPluginZookeeperEnable(c *gin.Context) {
	var req request.PluginsZookeeper
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := zookeeper.EnableZookeeperPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginZookeeperUpgrade 更新zookeeper插件信息
//
//	@Summary		更新zookeeper插件
//	@Description	更新zookeeper插件信息，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Zookeeper
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsZookeeper	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/zookeeper [put]
func ApiPluginZookeeperUpgrade(c *gin.Context) {
	var req request.PluginsZookeeper
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := zookeeper.UpgradeZookeeperPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginZookeeperGet 获取zookeeper插件信息
//
//	@Summary		获取zookeeper插件
//	@Description	获取zookeeper插件信息
//	@Schemes
//	@Tags		Plugin,Zookeeper
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/zookeeper [get]
func ApiPluginZookeeperGet(c *gin.Context) {
	resp, err := zookeeper.GetZookeeperPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
