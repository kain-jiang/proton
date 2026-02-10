package plugin

import (
	"net/http"

	"component-manage/internal/logic/mariadb"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/request"
	"component-manage/pkg/models/response"

	"github.com/gin-gonic/gin"
)

// ApiPluginMariaDBEnable 激活mariadb插件
//
//	@Summary		激活mariadb插件
//	@Description	激活mariadb插件，需要提供operater chart信息，镜像信息
//	@Schemes
//	@Tags		Plugin,MariaDB
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsMariaDB	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/mariadb [post]
func ApiPluginMariaDBEnable(c *gin.Context) {
	var req request.PluginsMariaDB
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := mariadb.EnableMariaDBPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginMariaDBEnable 更新mariadb插件
//
//	@Summary		更新mariadb插件
//	@Description	更新mariadb插件，需要提供operater chart信息，镜像信息
//	@Schemes
//	@Tags		Plugin,MariaDB
//	@Accept		json
//	@Produce	json
//	@Param		request	body		request.PluginsMariaDB	true	"Plugin Info"
//	@Success	200		{object}	response.Status
//	@Failure	400		{object}	cerr.E
//	@Failure	500		{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/mariadb [put]
func ApiPluginMariaDBUpgrade(c *gin.Context) {
	var req request.PluginsMariaDB
	if err := c.ShouldBindJSON(&req); err != nil {
		cerr.NewError(cerr.ParamsInvalidError, "params is invalied", err.Error()).Reply(c)
		return
	}

	if err := mariadb.UpgradeMariaDBPlugin(req); err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, response.OK)
}

// ApiPluginMariaDBGet 获取MariaDB插件信息
//
//	@Summary		获取MariaDB插件
//	@Description	获取MariaDB插件信息
//	@Schemes
//	@Tags		Plugin,MariaDB
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	response.Plugin
//	@Failure	400	{object}	cerr.E
//	@Failure	500	{object}	cerr.E
//	@Router		/api/component-manage/v1/components/plugin/mariadb [get]
func ApiPluginMariaDBGet(c *gin.Context) {
	resp, err := mariadb.GetMariaDBPlugin()
	if err != nil {
		cerr.AsError(err).Reply(c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
