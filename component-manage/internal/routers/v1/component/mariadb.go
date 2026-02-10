package component

import (
	"net/http"

	"component-manage/internal/logic/mariadb"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentKafkaCreate 创建mariadb
//
//	@Summary		创建mariadb
//	@Description	使用以前 proton-cli 的参数创建mariadb
//	@Schemes
//	@Tags		Component,MariaDB
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"MariaDB Component Name"
//	@Param		request	body		request.ComponentMariaDB	true	"MariaDB Create Params"
//	@Success	201		{object}	response.ComponentMariaDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mariadb/{name} [post]
func ApiComponentMariaDBCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentMariaDB
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := mariadb.CreateMariaDB(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentMariaDBUpgrade 更新mariadb
//
//	@Summary		更新mariadb
//	@Description	使用以前 proton-cli 的参数更新mariadb
//	@Schemes
//	@Tags		Component,MariaDB
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"Kafka Component Name"
//	@Param		request	body		request.ComponentMariaDB	true	"Kafka Update Params"
//	@Success	201		{object}	response.ComponentMariaDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mariadb/{name} [put]
func ApiComponentMariaDBUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentMariaDB
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := mariadb.UpgradeMariaDB(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentMariaDBGet 获取mariadb
//
//	@Summary		获取mariadb
//	@Description	获取mariadb，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,MariaDB
//	@Produce	json
//	@Param		name	path		string	true	"MariaDB Component Name"
//	@Success	200		{object}	response.ComponentMariaDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mariadb/{name} [get]
func ApiComponentMariaDBGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := mariadb.GetMariaDB(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentMariaDBList 获取所有mariadb
//
//	@Summary		获取所有mariadb
//	@Description	获取所有mariadb，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,MariaDB
//	@Produce	json
//	@Param		name	path		string	true	"MariaDB Component Name"
//	@Success	200		{object}	[]response.ComponentMariaDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mariadb [get]
func ApiComponentMariaDBList(c *gin.Context) {
	resp, err := mariadb.ListMariaDB()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentMariaDBDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := mariadb.DeleteMariaDB(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
