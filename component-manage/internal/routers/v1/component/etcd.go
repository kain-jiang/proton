package component

import (
	"net/http"

	"component-manage/internal/logic/etcd"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentOpensearchCreate 创建etcd
//
//	@Summary		创建etcd
//	@Description	使用以前 proton-cli 的参数创建etcd
//	@Schemes
//	@Tags		Component,ETCD
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"ETCD Component Name"
//	@Param		request	body		request.ComponentETCD	true	"ETCD Create Params"
//	@Success	201		{object}	response.ComponentETCD
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/etcd/{name} [post]
func ApiComponentETCDCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentETCD
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := etcd.CreateETCD(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentETCDUpgrade 更新etcd
//
//	@Summary		更新etcd
//	@Description	使用以前 proton-cli 的参数更新etcd
//	@Schemes
//	@Tags		Component,ETCD
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"ETCD Component Name"
//	@Param		request	body		request.ComponentETCD	true	"ETCD Create Params"
//	@Success	201		{object}	response.ComponentETCD
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/etcd/{name} [put]
func ApiComponentETCDUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentETCD
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := etcd.UpgradeETCD(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentETCDDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := etcd.DeleteETCD(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentETCDGet 获取etcd
//
//	@Summary		获取etcd
//	@Description	获取etcd，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,ETCD
//	@Produce	json
//	@Param		name	path		string	true	"ETCD Component Name"
//	@Success	200		{object}	response.ComponentETCD
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/etcd/{name} [get]
func ApiComponentETCDGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := etcd.GetETCD(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentOpensearchList 获取所有etcd
//
//	@Summary		获取所有etcd
//	@Description	获取所有etcd，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,ETCD
//	@Produce	json
//	@Success	200	{object}	[]response.ComponentOpensearch
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/etcd/ [get]
func ApiComponentETCDList(c *gin.Context) {
	resp, err := etcd.ListETCD()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
