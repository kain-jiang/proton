package plugin

import (
	"net/http"

	"component-manage/internal/logic/policyengine"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginPolicyEngineEnable 激活PolicyEngine插件
//
//	@Summary		激活PolicyEngine插件
//	@Description	激活PolicyEngine插件，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,PolicyEngine
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsPolicyEngine	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/policyengine [post]
func ApiPluginPolicyEngineEnable(c *gin.Context) {
	var req request.PluginsPolicyEngine
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := policyengine.EnablePolicyEnginePlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginPolicyEngineUpgrade 更新PolicyEngine插件信息
//
//	@Summary		更新PolicyEngine插件
//	@Description	更新PolicyEngine插件信息，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,PolicyEngine
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsPolicyEngine	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/policyengine [put]
func ApiPluginPolicyEngineUpgrade(c *gin.Context) {
	var req request.PluginsPolicyEngine
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := policyengine.UpgradePolicyEnginePlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginPolicyEngineGet 获取PolicyEngine插件信息
//
//	@Summary		获取PolicyEngine插件
//	@Description	获取PolicyEngine插件信息
//	@Schemes
//	@Tags		Plugin,PolicyEngine
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/policyengine [get]
func ApiPluginPolicyEngineGet(c *gin.Context) {
	resp, err := policyengine.GetPolicyEnginePlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
