package plugin

import (
	"net/http"

	"component-manage/internal/logic/etcd"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginETCDEnable 激活etcd插件
//
//	@Summary		激活etcd插件
//	@Description	激活etcd插件，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,ETCD
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsETCD	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/etcd [post]
func ApiPluginETCDEnable(c *gin.Context) {
	var req request.PluginsETCD
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := etcd.EnableETCDPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginETCDUpgrade 更新etcd插件信息
//
//	@Summary		更新etcd插件
//	@Description	更新etcd插件信息，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,ETCD
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsETCD	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/etcd [put]
func ApiPluginETCDUpgrade(c *gin.Context) {
	var req request.PluginsETCD
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := etcd.UpgradeETCDPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginETCDGet 获取etcd插件信息
//
//	@Summary		获取etcd插件
//	@Description	获取etcd插件信息
//	@Schemes
//	@Tags		Plugin,ETCD
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/etcd [get]
func ApiPluginETCDGet(c *gin.Context) {
	resp, err := etcd.GetETCDPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
