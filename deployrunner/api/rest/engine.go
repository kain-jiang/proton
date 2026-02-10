package rest

import (
	"net/http"
	"strconv"

	"taskrunner/pkg/app/executor"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"k8s.io/client-go/kubernetes"

	// load doc handler
	_ "taskrunner/api/rest/docs"
)

// ExecutorEngine wrapper app.TaskRunEngine interface into restful interface
type ExecutorEngine struct {
	*trait.System
	*executor.Executor
	*gin.Engine
	proxy   engineProxy
	router  *gin.RouterGroup
	irouter *gin.RouterGroup
}

// NewExecutorEngine return ExecutorEngine
func NewExecutorEngine(executor *executor.Executor, system trait.System) (*ExecutorEngine, *trait.Error) {
	e := &ExecutorEngine{
		Executor: executor,
		Engine:   gin.New(),
		System:   &system,
	}
	kcli, err := utils.NewKubeclient()
	if err != nil {
		return nil, err
	}
	err0 := e.RegistryHandler(system.NameSpace, kcli)
	return e, err0
}

func (e *ExecutorEngine) RouterGroups() []*gin.RouterGroup {
	return []*gin.RouterGroup{e.router, e.irouter}
}

func (e *ExecutorEngine) regitryHander(route *gin.RouterGroup) {
	{
		ga := route.Group("/application")
		ga.GET("/:aid", e.GetAPP)
		ga.PUT("/dependencesort", e.GetApplicationDependenceSort)
		ga.POST("/autodependence", e.GenerateAutoDependencies)
		ga.GET("/name/:name/:version", e.GetAPPWithName)
		ga.GET("/name/:name/", e.GetAPPWithName)
		ga.GET("", e.ListAPP)
		ga.POST("", e.UploadApplicationPackage)

		{
			gai := ga.Group("/instance/work")
			gai.GET("", e.ListWorkApplicationInstance)
		}

		{
			gc := ga.Group("/config")
			gc.GET("/:tid", e.GetConfigTemplate)
			gc.GET("", e.ListConfigTemplate)
			gc.POST("", e.UploadConfigTemplate)
			gc.DELETE("/:tid", e.DeleteConfigTemplate)
		}
	}
	{
		gs := route.Group("/system")
		gs.POST("", e.CreateSystemInfo)
		gs.PUT("", e.UpdateSystemInfo)
		gs.GET("", e.ListSystemInfo)
		gs.GET("/:sid", e.GetSystemInfo)
		gs.DELETE("/:sid", e.DeleteSystemInfo)

	}
	{
		gj := route.Group("/job")
		gj.POST("", e.CreateJob)
		gj.DELETE("", e.CreateDeleteAndStart)
		gj.GET("/:jid", e.GetJob)
		gj.DELETE("/:jid", e.MarkDeleteJobAndStart)

		gj.GET("/log", e.ListJobLog)
		gj.GET("", e.ListJob)
		gj.PUT("/:jid", e.SetJobConfig)
		gj.GET("/jsonschema/:jid", e.GetJobschema)
		gj.GET("/jsonschema/snapshot/:aid", e.GetSnapshotJobSchema)
		gj.GET("/jsonschema/snapshot/name/:name/:version", e.GetSnapshotJobSchemaWithName)
		gj.GET("/jsonschema/snapshot/name/:name/", e.GetSnapshotJobSchemaWithName)
		gj.POST("/jsonschema/snapshot", e.CreateAndSetJobConfig)

		// engine sepcial interface
		ge := gj.Group("/executor")
		ge.POST("/:jid", e.StartJob)
		ge.PATCH("/:jid", e.CanCelJob)

	}

	{
		route.GET("/log", e.ListJobLog)
	}
	{
		gdoc := route.Group("/openapi")
		gdoc.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	}
	{
		gc := route.Group("/component/instance")
		gc.GET(":cid", e.GetComponentInstance)
		gc.GET(":cid/dependence", e.ListComponentInstanceDependence)

	}

	// Update verification function related routing
	{
		gv := route.Group("/verification")
		gv.GET("/:jid", e.GetVerifyResult)
		gv.GET("/database", e.GetDataTestEntries)
		gv.GET("/function", e.GetFunctionTestEntries)
	}

	// Agents： images charts releases
	{
		ga := route.Group("/agents")
		ga.PUT("/image", e.UploadOCIImage)
		ga.PUT("/chart", e.UploadChart)
		ga.POST("/release/:name", e.InstallRelease)
		ga.DELETE("/release/:name", e.UninstallRelease)
	}
}

// RegistryHandler registry handler
func (e *ExecutorEngine) RegistryHandler(namespace string, kcli kubernetes.Interface) *trait.Error {
	prefix := "/api/deploy-installer/v1"
	e.irouter = e.Engine.Group("/internal" + prefix)
	e.router = e.Engine.Group(prefix)
	if namespace != "" {
		e.router.Use(newOauthMiddleware(e.Log, namespace, kcli).Authentication)
	}

	e.regitryHander(e.irouter)
	e.regitryHander(e.router)

	e.proxy.host = "http://" + "deploy-installer-%d:9090/internal" + prefix

	return nil
}

// StartJob start the job
//
//	@Summary		执行任务
//	@Description	执行已确认的任务,开始后台执行安装/更新任务。
//	@Description	不允许重复执行正在执行中的任务
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			jid						path		int			true	"任务ID"
//
//	@Success		200						{object}	int			"返回为空"
//	@Failure		400						{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		412						{object}	HTTPError	"任务配置未确认,不允许启动"
//	@Failure		500						{object}	HTTPError	"系统内部错误"
//	@Router			/job/executor/{jid} 	[post]
func (e *ExecutorEngine) StartJob(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, err0 := strconv.Atoi(jidStr)
	if err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	err := e.Executor.StartJob(ctx, jid)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrConfigNotComfirm) {
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrJobExecuting) {
			// job has been executor, no error
			ctx.JSON(http.StatusOK, nil)
		} else if trait.IsInternalError(err, trait.ErrJobOwnerError) {
			// job has been executor by other executor, no error
			e.proxy.StartJob(ctx, jid, err.Detail.(int))
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, nil)
	}
}

// CanCelJob stop the job
//
//	@Summary		暂停执行中的任务
//	@Description	暂停执行中的任务，由于已通过安装阶段的任务暂停可能会造成较大的影响，因此不允许暂停该阶段往后的任务
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			jid						path		int			true	"任务ID"
//
//	@Success		200						{object}	int			"返回为空"
//	@Failure		400						{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		412						{object}	HTTPError	"不允许设置任务配置"
//	@Failure		500						{object}	HTTPError	"系统内部错误"
//	@Router			/job/executor/{jid} 	[patch]
func (e *ExecutorEngine) CanCelJob(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, err0 := strconv.Atoi(jidStr)
	if err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	err := e.Executor.CancelJob(ctx, jid)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrJobCantStop) {
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrJobOwnerError) {
			// TODO proxy this request
			e.proxy.CancelJob(ctx, jid, err.Detail.(int))
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, nil)
	}
}
