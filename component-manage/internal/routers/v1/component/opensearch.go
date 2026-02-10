package component

import (
	"net/http"

	"component-manage/internal/logic/opensearch"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentOpensearchCreate 创建opensearch
//
//	@Summary		创建opensearch
//	@Description	使用以前 proton-cli 的参数创建opensearch
//	@Schemes
//	@Tags		Component,OpenSearch
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"Opensearch Component Name"
//	@Param		request	body		request.ComponentOpensearch	true	"Opensearch Create Params"
//	@Success	201		{object}	response.ComponentOpensearch
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/opensearch/{name} [post]
func ApiComponentOpensearchCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentOpensearch
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := opensearch.CreateOpensearch(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentOpensearchUpgrade 更新opensearch
//
//	@Summary		更新opensearch
//	@Description	使用以前 proton-cli 的参数更新opensearch
//	@Schemes
//	@Tags		Component,OpenSearch
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"Opensearch Component Name"
//	@Param		request	body		request.ComponentOpensearch	true	"Opensearch Create Params"
//	@Success	201		{object}	response.ComponentOpensearch
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/opensearch/{name} [put]
func ApiComponentOpensearchUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentOpensearch
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := opensearch.UpgradeOpensearch(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentOpensearchDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := opensearch.DeleteOpensearch(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentOpensearchGet 获取opensearch
//
//	@Summary		获取opensearch
//	@Description	获取opensearch，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,OpenSearch
//	@Produce	json
//	@Param		name	path		string	true	"Opensearch Component Name"
//	@Success	200		{object}	response.ComponentOpensearch
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/opensearch/{name} [get]
func ApiComponentOpensearchGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := opensearch.GetOpensearch(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentOpensearchList 获取所有kafka
//
//	@Summary		获取所有opensearch
//	@Description	获取所有opensearch，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,OpenSearch
//	@Produce	json
//	@Success	200	{object}	[]response.ComponentOpensearch
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/opensearch/ [get]
func ApiComponentOpensearchList(c *gin.Context) {
	resp, err := opensearch.ListOpensearch()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
