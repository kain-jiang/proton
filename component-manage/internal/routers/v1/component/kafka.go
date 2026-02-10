package component

import (
	"net/http"

	"component-manage/internal/logic/kafka"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentKafkaCreate 创建kafka
//
//	@Summary		创建kafka
//	@Description	使用以前 proton-cli 的参数创建kafka
//	@Schemes
//	@Tags		Component,Kafka
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"Kafka Component Name"
//	@Param		request	body		request.ComponentKafka	true	"Kafka Create Params"
//	@Success	201		{object}	response.ComponentKafka
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/kafka/{name} [post]
func ApiComponentKafkaCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentKafka
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := kafka.CreateKafka(name, param.Params, param.Dependencies.Zookeeper)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentKafkaUpgrade 更新kafka
//
//	@Summary		更新kafka
//	@Description	使用以前 proton-cli 的参数更新kafka
//	@Schemes
//	@Tags		Component,Kafka
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string					true	"Kafka Component Name"
//	@Param		request	body		request.ComponentKafka	true	"Kafka Update Params"
//	@Success	201		{object}	response.ComponentKafka
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/kafka/{name} [put]
func ApiComponentKafkaUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentKafka
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := kafka.UpgradeKafka(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentKafkaDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := kafka.DeleteKafka(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentKafkaGet 获取kafka
//
//	@Summary		获取kafka
//	@Description	获取kafka，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Kafka
//	@Produce	json
//	@Param		name	path		string	true	"Kafka Component Name"
//	@Success	200		{object}	response.ComponentKafka
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/kafka/{name} [get]
func ApiComponentKafkaGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := kafka.GetKafka(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentKafkaList 获取所有kafka
//
//	@Summary		获取所有kafka
//	@Description	获取所有kafka，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Kafka
//	@Produce	json
//	@Success	200	{object}	[]response.ComponentKafka
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/kafka/ [get]
func ApiComponentKafkaList(c *gin.Context) {
	resp, err := kafka.ListKafka()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
