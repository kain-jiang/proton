package rest

import (
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type (
	respUploadOCIImage struct {
		Images []map[string]string `json:"images"`
	}
)

// UploadOCIImage upload a oci image archive
//
//	@Summary		代理实现上传一个oci镜像包
//	@Description	在当前的CR里推送此镜像包，并返回镜像基本信息
//
//	@Tags			agents
//	@Accept			application/octet-stream
//	@Produce		json
//
//	@Success		200				{object}	respUploadOCIImage	"镜像信息"
//	@Failure		500				{object}	HTTPError			"系统内部错误"
//	@Router			/agents/image 	[put]
func (e *ExecutorEngine) UploadOCIImage(ctx *gin.Context) {
	tmpfile, err := os.CreateTemp("", "*.tar")
	if err != nil {
		e.Log.WithError(err).Error("agents/image: create temp file failed")
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	defer os.Remove(tmpfile.Name())

	_, err = io.Copy(tmpfile, ctx.Request.Body)
	if err != nil {
		e.Log.WithError(err).Error("agents/image: copy request body to tmpfile failed")
		ParamError.From(err.Error()).AbortGin(ctx)
		return

	}
	tmpfile.Seek(0, 0)

	imgs, terr := e.Executor.PushImages(ctx, tmpfile.Name())
	if terr != nil {
		e.Log.WithError(terr).Error("agents/image: push images failed")
		UnknownError.From(terr.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, respUploadOCIImage{
		Images: imgs,
	})
}

type (
	respUploadChart struct {
		Chart  map[string]any `json:"chart"`
		Values map[string]any `json:"values"`
	}
)

// UploadChart upload a chart package
//
//	@Summary		代理实现上传一个chart包
//	@Description	在当前的CR里推送此chart包，并返回Chart基本信息
//
//	@Tags			agents
//	@Accept			application/octet-stream
//	@Produce		json
//
//	@Success		200				{object}	respUploadChart	"Chart信息"
//	@Failure		500				{object}	HTTPError		"系统内部错误"
//	@Router			/agents/chart 	[put]
func (e *ExecutorEngine) UploadChart(ctx *gin.Context) {
	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		e.Log.WithError(err).Error("agents/chart: read request body failed")
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	cht, val, terr := e.Executor.PushChart(ctx, data)
	if terr != nil {
		e.Log.WithError(terr).Error("agents/chart: push chart failed")
		UnknownError.From(terr.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, respUploadChart{
		Chart:  cht,
		Values: val,
	})
}

type (
	respInstallRelease struct {
		Values map[string]any `json:"values"`
	}
	reqInstallRelease struct {
		Name    string         `json:"name" binding:"required"`
		Version string         `json:"version" binding:"required"`
		Values  map[string]any `json:"values" binding:"required"`
	}
)

// InstallRelease install a helm release
//
//	@Summary		代理实现安装一个helm chart
//	@Description	在当前的系统安装/更新一个chart，并返回创建的release信息
//
//	@Tags			agents
//	@Accept			json
//	@Produce		json
//
//	@Param			name					path		string				true	"Release名"
//	@Param			namespace				query		string				true	"命名空间"
//	@Param			set-registry			query		boolean				false	"设置仓库地址，默认设置"
//	@Param			request					body		reqInstallRelease	true	"安装信息信息"
//
//	@Success		200						{object}	respInstallRelease	"Release信息"
//	@Failure		500						{object}	HTTPError			"系统内部错误"
//	@Router			/agents/release/:name 	[put]
func (e *ExecutorEngine) InstallRelease(ctx *gin.Context) {
	var req reqInstallRelease
	var qry struct {
		Namespace   string `form:"namespace" binding:"required"`
		SetRegistry bool   `form:"set-registry,default=true"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		e.Log.WithError(err).Error("agents/release: bind request body failed")
		ParamError.From(err.Error()).AbortGin(ctx)
		return
	}

	if err := ctx.ShouldBindQuery(&qry); err != nil {
		e.Log.WithError(err).Error("agents/release: bind query failed")
		ParamError.From(err.Error()).AbortGin(ctx)
		return
	}

	name := ctx.Param("name")
	if name == "" {
		ParamError.From("path param name is required").AbortGin(ctx)
		return
	}

	vals, terr := e.Executor.InstallRelease(
		ctx,
		name,
		req.Name,
		req.Version,
		qry.Namespace,
		qry.SetRegistry,
		req.Values,
	)
	if terr != nil {
		e.Log.WithError(terr).Error("agents/release: install release failed")
		UnknownError.From(terr.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, respInstallRelease{Values: vals})
}

type (
	respUninstallRelease struct {
		Values map[string]any `json:"values"`
	}
)

// UninstallRelease uninstall a release
//
//	@Summary		代理实现卸载一个helm实例
//	@Description	在当前的系统卸载一个release
//
//	@Tags			agents
//	@Produce		json
//
//	@Param			name					path		string					true	"Release名"
//	@Param			namespace				query		string					true	"命名空间"
//
//	@Success		200						{object}	respUninstallRelease	"Release信息"
//	@Failure		500						{object}	HTTPError				"系统内部错误"
//	@Router			/agents/release/{name} 	[put]
func (e *ExecutorEngine) UninstallRelease(ctx *gin.Context) {
	var qry struct {
		Namespace string `form:"namespace" binding:"required"`
	}

	if err := ctx.ShouldBindQuery(&qry); err != nil {
		e.Log.WithError(err).Error("agents/release: bind query failed")
		ParamError.From(err.Error()).AbortGin(ctx)
		return
	}

	name := ctx.Param("name")
	if name == "" {
		ParamError.From("path param name is required").AbortGin(ctx)
		return
	}

	values, terr := e.Executor.UninstallRelease(
		ctx,
		name,
		qry.Namespace,
	)
	if terr != nil {
		e.Log.WithError(terr).Error("agents/release: uninstall release failed")
		UnknownError.From(terr.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, respUninstallRelease{Values: values})
}
