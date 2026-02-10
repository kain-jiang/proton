package plugin

import (
	"net/http"

	"component-manage/internal/logic/opensearch"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginOpensearchEnable 激活opensearch插件
//
//	@Summary		激活opensearch插件
//	@Description	激活opensearch插件，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,OpenSearch
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsOpensearch	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/opensearch [post]
func ApiPluginOpensearchEnable(c *gin.Context) {
	var req request.PluginsOpensearch
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := opensearch.EnableOpensearchPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginOpensearchUpgrade 更新opensearch插件信息
//
//	@Summary		更新opensearch插件
//	@Description	更新opensearch插件信息，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,OpenSearch
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsOpensearch	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/opensearch [put]
func ApiPluginOpensearchUpgrade(c *gin.Context) {
	var req request.PluginsOpensearch
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := opensearch.UpgradeOpensearchPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginOpensearchGet 获取opensearch插件信息
//
//	@Summary		获取opensearch插件
//	@Description	获取opensearch插件信息
//	@Schemes
//	@Tags		Plugin,OpenSearch
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/opensearch [get]
func ApiPluginOpensearchGet(c *gin.Context) {
	resp, err := opensearch.GetOpensearchPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
