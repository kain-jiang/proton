package plugin

import (
	"net/http"

	"component-manage/internal/logic/mongodb"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginMongoDBEnable 激活mongodb插件
//
//	@Summary		激活mongodb插件
//	@Description	激活mongodb插件，需要提供operater chart信息，镜像信息
//	@Schemes
//	@Tags		Plugin,MongoDB
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsMongoDB	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/mongodb [post]
func ApiPluginMongoDBEnable(c *gin.Context) {
	var req request.PluginsMongoDB
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := mongodb.EnableMongoDBPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginMongoDBEnable 更新mongodb插件
//
//	@Summary		更新mongodb插件
//	@Description	更新mongodb插件，需要提供operater chart信息，镜像信息
//	@Schemes
//	@Tags		Plugin,MongoDB
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsMongoDB	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/mongodb [put]
func ApiPluginMongoDBUpgrade(c *gin.Context) {
	var req request.PluginsMongoDB
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := mongodb.UpgradeMongoDBPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginMongoDBGet 获取MongoDB插件信息
//
//	@Summary		获取MongoDB插件
//	@Description	获取MongoDB插件信息
//	@Schemes
//	@Tags		Plugin,MongoDB
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/mongodb [get]
func ApiPluginMongoDBGet(c *gin.Context) {
	resp, err := mongodb.GetMongoDBPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
