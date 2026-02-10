package plugin

import (
	"net/http"

	"component-manage/internal/logic/kafka"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginKafkaEnable 激活Kafka插件
//
//	@Summary		激活Kafka插件
//	@Description	激活Kafka插件，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Kafka
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsKafka	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/kafka [post]
func ApiPluginKafkaEnable(c *gin.Context) {
	var req request.PluginsKafka
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := kafka.EnableKafkaPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginKafkaUpgrade 更新Kafka插件信息
//
//	@Summary		更新Kafka插件
//	@Description	更新Kafka插件信息，需要提供chart信息
//	@Schemes
//	@Tags		Plugin,Kafka
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsKafka	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/kafka [put]
func ApiPluginKafkaUpgrade(c *gin.Context) {
	var req request.PluginsKafka
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := kafka.UpgradeKafkaPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginKafkaGet 获取Kafka插件信息
//
//	@Summary		获取Kafka插件
//	@Description	获取Kafka插件信息
//	@Schemes
//	@Tags		Plugin,Kafka
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/kafka [get]
func ApiPluginKafkaGet(c *gin.Context) {
	resp, err := kafka.GetKafkaPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
