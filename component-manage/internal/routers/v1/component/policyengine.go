package component

import (
	"net/http"

	"component-manage/internal/logic/policyengine"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentPolicyEngineCreate 创建policyengine
//
//	@Summary		创建policyengine
//	@Description	使用以前 proton-cli 的参数创建policyengine
//	@Schemes
//	@Tags		Component,PolicyEngine
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string							true	"PolicyEngine Component Name"
//	@Param		request	body		request.ComponentPolicyEngine	true	"PolicyEngine Create Params"
//	@Success	201		{object}	response.ComponentPolicyEngine
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/policyengine/{name} [post]
func ApiComponentPolicyEngineCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentPolicyEngine
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := policyengine.CreatePolicyEngine(name, param.Params, param.Dependencies.ETCD)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentPolicyEngineUpgrade 更新policyengine
//
//	@Summary		更新policyengine
//	@Description	使用以前 proton-cli 的参数更新policyengine
//	@Schemes
//	@Tags		Component,PolicyEngine
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string							true	"PolicyEngine Component Name"
//	@Param		request	body		request.ComponentPolicyEngine	true	"PolicyEngine Create Params"
//	@Success	201		{object}	response.ComponentPolicyEngine
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/policyengine/{name} [put]
func ApiComponentPolicyEngineUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentPolicyEngine
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := policyengine.UpgradePolicyEngine(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentPolicyEngineDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := policyengine.DeletePolicyEngine(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentPolicyEngineGet 获取policyengine
//
//	@Summary		获取policyengine
//	@Description	获取policyengine，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,PolicyEngine
//	@Produce	json
//	@Param		name	path		string	true	"PolicyEngine Component Name"
//	@Success	200		{object}	response.ComponentPolicyEngine
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/policyengine/{name} [get]
func ApiComponentPolicyEngineGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := policyengine.GetPolicyEngine(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentPolicyEngineList 获取所有policyengine
//
//	@Summary		获取所有policyengine
//	@Description	获取所有policyengine，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,PolicyEngine
//	@Produce	json
//	@Success	200	{object}	[]response.ComponentPolicyEngine
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/policyengine/ [get]
func ApiComponentPolicyEngineList(c *gin.Context) {
	resp, err := policyengine.ListPolicyEngine()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
