package component

import (
	"net/http"

	"component-manage/internal/logic/nebula"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentNebulaCreate 创建nebula
//
//	@Summary		创建nebula
//	@Description	使用以前 proton-cli 的参数创建nebula
//	@Schemes
//	@Tags		Component,Nebula
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"Nebula Component Name"
//	@Param		request	body		request.ComponentNebula	true	"Nebula Create Params"
//	@Success	201		{object}	response.ComponentNebula
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/nebula/{name} [post]
func ApiComponentNebulaCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentNebula
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := nebula.CreateNebula(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func ApiComponentNebulaDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := nebula.DeleteNebula(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentNebulaGet 获取nebula
//
//	@Summary		获取nebula
//	@Description	获取nebula，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Nebula
//	@Produce	json
//	@Param		name	path		string	true	"Nebula Component Name"
//	@Success	200		{object}	response.ComponentNebula
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/nebula/{name} [get]
func ApiComponentNebulaGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := nebula.GetNebula(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentNebulaList 获取所有nebula
//
//	@Summary		获取所有nebula
//	@Description	获取所有nebula，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Nebula
//	@Produce	json
//	@Param		name	path		string	true	"Nebula Component Name"
//	@Success	200		{object}	[]response.ComponentNebula
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/nebula [get]
func ApiComponentNebulaList(c *gin.Context) {
	resp, err := nebula.ListNebula()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentNebulaUpgrade 更新nebula
//
//	@Summary		更新nebula
//	@Description	使用以前 proton-cli 的参数更新nebula
//	@Schemes
//	@Tags		Component,Nebula
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"Nebula Component Name"
//	@Param		request	body		request.ComponentNebula	true	"Nebula Update Params"
//	@Success	201		{object}	response.ComponentNebula
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/nebula/{name} [put]
func ApiComponentNebulaUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentNebula
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := nebula.UpgradeNebula(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
