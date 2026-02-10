package component

import (
	"net/http"

	"component-manage/internal/logic/redis"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentRedisCreate 创建redis
//
//	@Summary		创建redis
//	@Description	使用以前 proton-cli 的参数创建redis
//	@Schemes
//	@Tags		Component,Redis
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"Redis Component Name"
//	@Param		request	body		request.ComponentRedis	true	"Redis Create Params"
//	@Success	201		{object}	response.ComponentRedis
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/redis/{name} [post]
func ApiComponentRedisCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentRedis
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := redis.CreateRedis(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentRedisUpgrade 更新redis
//
//	@Summary		更新redis
//	@Description	使用以前 proton-cli 的参数更新redis
//	@Schemes
//	@Tags		Component,Redis
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"Redis Component Name"
//	@Param		request	body		request.ComponentRedis	true	"Redis Create Params"
//	@Success	201		{object}	response.ComponentRedis
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/redis/{name} [put]
func ApiComponentRedisUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentRedis
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := redis.UpgradeRedis(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentRedisDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := redis.DeleteRedis(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentRedisGet 获取redis
//
//	@Summary		获取redis
//	@Description	获取redis，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Redis
//	@Produce	json
//	@Param		name	path		string	true	"Redis Component Name"
//	@Success	200		{object}	response.ComponentRedis
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/redis/{name} [get]
func ApiComponentRedisGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := redis.GetRedis(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentOpensearchList 获取所有redis
//
//	@Summary		获取所有redis
//	@Description	获取所有redis，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Redis
//	@Produce	json
//	@Success	200	{object}	[]response.ComponentOpensearch
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/redis/ [get]
func ApiComponentRedisList(c *gin.Context) {
	resp, err := redis.ListRedis()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
