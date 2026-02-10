package component

import (
	"net/http"

	"component-manage/internal/logic/mongodb"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentMongoDBCreate 创建mongodb
//
//	@Summary		创建mongodb
//	@Description	使用以前 proton-cli 的参数创建mongodb
//	@Schemes
//	@Tags		Component,MongoDB
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"MongoDB Component Name"
//	@Param		request	body		request.ComponentMongoDB	true	"MongoDB Create Params"
//	@Success	201		{object}	response.ComponentMongoDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mongodb/{name} [post]
func ApiComponentMongoDBCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentMongoDB
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := mongodb.CreateMongoDB(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func ApiComponentMongoDBDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := mongodb.DeleteMongoDB(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentMongoDBGet 获取mongodb
//
//	@Summary		获取mongodb
//	@Description	获取mongodb，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,MongoDB
//	@Produce	json
//	@Param		name	path		string	true	"MongoDB Component Name"
//	@Success	200		{object}	response.ComponentMongoDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mongodb/{name} [get]
func ApiComponentMongoDBGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := mongodb.GetMongoDB(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentMongoDBList 获取所有mongodb
//
//	@Summary		获取所有mongodb
//	@Description	获取所有mongodb，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,MongoDB
//	@Produce	json
//	@Param		name	path		string	true	"MongoDB Component Name"
//	@Success	200		{object}	[]response.ComponentMongoDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mongodb [get]
func ApiComponentMongoDBList(c *gin.Context) {
	resp, err := mongodb.ListMongoDB()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentMongoDBUpgrade 更新mongodb
//
//	@Summary		更新mongodb
//	@Description	使用以前 proton-cli 的参数更新mongodb
//	@Schemes
//	@Tags		Component,MongoDB
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"MongoDB Component Name"
//	@Param		request	body		request.ComponentMongoDB	true	"MongoDB Update Params"
//	@Success	201		{object}	response.ComponentMongoDB
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/mongodb/{name} [put]
func ApiComponentMongoDBUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentMongoDB
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := mongodb.UpgradeMongoDB(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
