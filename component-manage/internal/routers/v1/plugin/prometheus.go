package plugin

import (
	"net/http"

	"component-manage/internal/logic/prometheus"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginPrometheusEnable 激活prometheus插件
//
//	@Summary		激活prometheus插件
//	@Description	激活prometheus插件，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Prometheus
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsPrometheus	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/prometheus [post]
func ApiPluginPrometheusEnable(c *gin.Context) {
	var req request.PluginsPrometheus
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := prometheus.EnablePrometheusPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginPrometheusUpgrade 更新prometheus插件信息
//
//	@Summary		更新prometheus插件
//	@Description	更新prometheus插件信息，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Prometheus
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsPrometheus	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/prometheus [put]
func ApiPluginPrometheusUpgrade(c *gin.Context) {
	var req request.PluginsPrometheus
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := prometheus.UpgradePrometheusPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginPrometheusGet 获取prometheus插件信息
//
//	@Summary		获取prometheus插件
//	@Description	获取prometheus插件信息
//	@Schemes
//	@Tags		Plugin,Prometheus
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/prometheus [get]
func ApiPluginPrometheusGet(c *gin.Context) {
	resp, err := prometheus.GetPrometheusPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
