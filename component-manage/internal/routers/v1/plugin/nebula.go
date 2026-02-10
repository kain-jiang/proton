package plugin

import (
	"net/http"

	"component-manage/internal/logic/nebula"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginNebulaEnable 激活nebula插件
//
//	@Summary		激活nebula插件
//	@Description	激活nebula插件，需要提供operater chart信息，镜像信息
//	@Schemes
//	@Tags		Plugin,Nebula
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsNebula	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/nebula [post]
func ApiPluginNebulaEnable(c *gin.Context) {
	var req request.PluginsNebula
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := nebula.EnableNebulaPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginNebulaEnable 更新nebula插件
//
//	@Summary		更新nebula插件
//	@Description	更新nebula插件，需要提供operater chart信息，镜像信息
//	@Schemes
//	@Tags		Plugin,Nebula
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsNebula	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/nebula [put]
func ApiPluginNebulaUpgrade(c *gin.Context) {
	var req request.PluginsNebula
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := nebula.UpgradeNebulaPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginNebulaGet 获取Nebula插件信息
//
//	@Summary		获取Nebula插件
//	@Description	获取Nebula插件信息
//	@Schemes
//	@Tags		Plugin,Nebula
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/nebula [get]
func ApiPluginNebulaGet(c *gin.Context) {
	resp, err := nebula.GetNebulaPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
