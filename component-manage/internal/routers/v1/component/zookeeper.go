package component

import (
	"net/http"

	"component-manage/internal/logic/zookeeper"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	_ "component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiComponentZookeeperCreate 创建zookeeper
//
//	@Summary		创建zookeeper
//	@Description	使用以前 proton-cli 的参数创建zookeeper
//	@Schemes
//	@Tags		Component,Zookeeper
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"Zookeeper Component Name"
//	@Param		request	body		request.ComponentZookeeper	true	"Zookeeper Create Params"
//	@Success	201		{object}	response.ComponentZookeeper
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/zookeeper/{name} [post]
func ApiComponentZookeeperCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentZookeeper
	if err := c.ShouldBindJSON(&param); err != nil {
		(&cerr.E{
			Code:    cerr.ParamsInvalidError,
			Message: "params is invalid",
			Cause:   err.Error(),
		}).Reply(c)
		return
	}

	resp, err := zookeeper.CreateZookeeper(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentZookeeperUpgrade 更新zookeeper
//
//	@Summary		更新zookeeper
//	@Description	使用以前 proton-cli 的参数更新zookeeper
//	@Schemes
//	@Tags		Component,Zookeeper
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"Zookeeper Component Name"
//	@Param		request	body		request.ComponentZookeeper	true	"Zookeeper Update Params"
//	@Success	200		{object}	response.ComponentZookeeper
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/zookeeper/{name} [put]
func ApiComponentZookeeperUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentZookeeper
	if err := c.ShouldBindJSON(&param); err != nil {
		(&cerr.E{
			Code:    cerr.ParamsInvalidError,
			Message: "params is invalid",
			Cause:   err.Error(),
		}).Reply(c)
		return
	}

	resp, err := zookeeper.UpgradeZookeeper(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentZookeeperDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := zookeeper.DeleteZookeeper(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentZookeeperGet 获取zookeeper
//
//	@Summary		获取zookeeper
//	@Description	获取zookeeper，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Zookeeper
//	@Produce	json
//	@Param		name	path		string	true	"Zookeeper Component Name"
//	@Success	200		{object}	response.ComponentZookeeper
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/zookeeper/{name} [get]
func ApiComponentZookeeperGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := zookeeper.GetZookeeper(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentZookeeperList 获取所有zookeeper
//
//	@Summary		获取所有zookeeper
//	@Description	获取所有zookeeper，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Zookeeper
//	@Produce	json
//	@Success	200	{object}	[]response.ComponentZookeeper
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/zookeeper [get]
func ApiComponentZookeeperList(c *gin.Context) {
	resp, err := zookeeper.ListZookeeper()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
