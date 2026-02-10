package component

import (
	"net/http"

	"component-manage/internal/logic/prometheus"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"

	"github.com/gin-gonic/gin"
)

// ApiComponentPrometheusCreate 创建prometheus
//
//	@Summary		创建prometheus
//	@Description	使用以前 proton-cli 的参数创建prometheus
//	@Schemes
//	@Tags		Component,Prometheus
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"Prometheus Component Name"
//	@Param		request	body		request.ComponentPrometheus	true	"Prometheus Create Params"
//	@Success	201		{object}	response.ComponentPrometheus
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/prometheus/{name} [post]
func ApiComponentPrometheusCreate(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentPrometheus
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := prometheus.CreatePrometheus(name, param.Params, param.Dependencies.ETCD)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// ApiComponentPrometheusUpgrade 更新prometheus
//
//	@Summary		更新prometheus
//	@Description	使用以前 proton-cli 的参数更新prometheus
//	@Schemes
//	@Tags		Component,Prometheus
//	@Accept		json
//	@Produce	json
//	@Param		name	path		string						true	"Prometheus Component Name"
//	@Param		request	body		request.ComponentPrometheus	true	"Prometheus Create Params"
//	@Success	201		{object}	response.ComponentPrometheus
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/prometheus/{name} [put]
func ApiComponentPrometheusUpgrade(c *gin.Context) {
	name := c.Param("name")

	var param request.ComponentPrometheus
	if err := c.ShouldBindJSON(&param); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "param is invalid", err.Error()).Reply(c)
		return
	}

	resp, err := prometheus.UpgradePrometheus(name, param.Params)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func ApiComponentPrometheusDelete(c *gin.Context) {
	name := c.Param("name")
	_, toClean := c.GetQuery("clean")

	resp, err := prometheus.DeletePrometheus(name, toClean)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentPrometheusGet 获取prometheus
//
//	@Summary		获取prometheus
//	@Description	获取prometheus，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Prometheus
//	@Produce	json
//	@Param		name	path		string	true	"Prometheus Component Name"
//	@Success	200		{object}	response.ComponentPrometheus
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/prometheus/{name} [get]
func ApiComponentPrometheusGet(c *gin.Context) {
	name := c.Param("name")

	resp, err := prometheus.GetPrometheus(name)
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ApiComponentOpensearchList 获取所有prometheus
//
//	@Summary		获取所有prometheus
//	@Description	获取所有prometheus，得到proton-cli需要的连接信息
//	@Schemes
//	@Tags		Component,Prometheus
//	@Produce	json
//	@Success	200	{object}	[]response.ComponentOpensearch
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/release/prometheus/ [get]
func ApiComponentPrometheusList(c *gin.Context) {
	resp, err := prometheus.ListPrometheus()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
