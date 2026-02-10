package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"taskrunner/error/codes"
	"taskrunner/pkg/app"
	"taskrunner/pkg/app/builder"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
)

// aidBody only be use to create api doc
//
//nolint:unused
type aidBody struct {
	// 应用包ID
	AID int `json:"aid"`
	// system id，系统为多应用实例隔离空间
	Sid int `json:"sid"`
}

// CreateJob create the application's job
//
//	@Summary		在系统同创建一个应用的安装或更新准备任务。
//	@Description	以当前系统正在工作的应用为当前配置,创建一个安装或更新准备任务。
//	@Description	可以通过准备任务接口执行该准备任务,初次创建的任务需要进行配置确认,
//	@Description	配置确认的目的在于确定升级后不会导致配置丢失以及配置变更后没有处理等问题。
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			aid		body		aidBody		true	"需要安装应用id"
//
//	@Success		200		{object}	int			"任务ID"
//	@Failure		412		{object}	HTTPError	"客户端请求参数错误,code为412017004时表示组件依赖缺少"
//	@Failure		409		{object}	HTTPError	"客户端请求重复创建"
//	@Failure		500		{object}	HTTPError	"系统内部错误"
//	@Router			/job 	[post]
func (e *ExecutorEngine) CreateJob(ctx *gin.Context) {
	meta := &aidBody{
		AID: -1,
	}
	if err0 := ctx.BindJSON(meta); err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	sid := meta.Sid
	if e.SID >= 0 {
		sid = e.SID
	}
	if meta.AID == -1 {
		ParamError.From("application aid must set").AbortGin(ctx)
		return
	}

	app, err := e.Store.GetAPP(ctx, meta.AID)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	err = e.CheckDepComponents(ctx, app, sid)
	if trait.IsInternalError(err, trait.ErrComponentNotFound) {
		ComponentNotfoundError.From(err.ToJson()).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	id, err := e.Store.NewJobRecord(ctx, meta.AID, sid)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrUniqueKey) {
			UniqueKeyError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, id)
	}
}

// CreateAndSetJobConfig create a job and set the job with config then start it
//
//	@Summary		以指定配置创建任务并启动
//	@Description	创建aid对应的job，并以指定的配置设置job，并启动对应任务
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			config	body		jobSChemaConfig	true	"目标应用配置"
//
//	@Success		200		{object}	int				"返回任务jid"
//	@Failure		400		{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404		{object}	HTTPError		"对象不存在，不允许更新"
//	@Failure		412		{object}	HTTPError		"不允许设置任务配置"
//	@Failure		500		{object}	HTTPError		"系统内部错误"
//	@Router			/job/jsonschema/snapshot [post]
func (e *ExecutorEngine) CreateAndSetJobConfig(ctx *gin.Context) {
	a := &jobSChemaConfig{}
	if err0 := ctx.BindJSON(a); err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	sid := a.SID
	if e.SID >= 0 {
		sid = e.SID
	}
	ains := a.ToAppIns()
	ains.AID = a.AID
	ains.SID = sid
	jid, err := e.CreateAndStartJobWithConfig(ctx, ains)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrComponentNotFound) {
			ComponentNotfoundError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrJobExecuting) {
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, app.ErrNoAvailableWorker) {
			err.Detail = fmt.Errorf("job %d has been create, but %s, please try exeute it later", jid, err.Error())
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrConfigValidate) {
			ParamError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, jid)
	}
}

type ApplicationinstanceMeta struct {
	// 系统空间
	SID int `json:"sid,omitempty"`
	// 应用名称
	AName string `json:"name"`
}

// CreateDeleteAndStart create a delete job then start it
//
//	@Summary		删除当前系统指定应用
//	@Description	创建当前系统指定应用删除任务,并启动对应任务
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			sid		query		int			false	"系统空间ID,多实例下必填"
//	@Param			force	query		bool		false	"忽略依赖检查,强制卸载,默认检查"
//	@Param			name	query		string		true	"应用标识名称"
//
//	@Success		200		{object}	int			"返回任务jid"
//	@Failure		400		{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404		{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		412		{object}	HTTPError	"不允许设置任务配置"
//	@Failure		500		{object}	HTTPError	"系统内部错误"
//	@Router			/job [DELETE]
func (e *ExecutorEngine) CreateDeleteAndStart(ctx *gin.Context) {
	a := &ApplicationinstanceMeta{}
	var sid int
	if e.SID >= 0 {
		sid = e.SID
	} else {
		sidQuery, err := parseIntFromQuery(ctx, "sid")
		if err != nil {
			err.AbortGin(ctx)
		}
		sid = sidQuery[0]
	}
	a.SID = sid
	a.AName = ctx.Query("name")
	if a.AName == "" {
		ParamError.From("name must not empty").AbortGin(ctx)
		return
	}
	ains, err := e.GetWorkAPPIns(ctx, a.AName, a.SID)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			err.Detail = fmt.Sprintf("the application %s no need delete", a.AName)
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
		return
	}
	if force := ctx.Query("force"); force != "true" {
		index := map[int][]int{}
		for _, c := range ains.Components {
			froms, err := e.Store.GetPointTo(ctx, c.CID)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			index[c.CID] = froms
		}
		outer := hasOuterEdge(index)
		edges := make([]trait.Edge, 0, len(outer))
		for _, edge := range outer {
			from, err := e.Store.GetComponentIns(ctx, edge[0])
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			to, err := e.Store.GetComponentIns(ctx, edge[1])
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			edges = append(edges, trait.Edge{
				From: from.Component,
				To:   to.Component,
			})
		}
		if len(edges) != 0 {
			err := trait.Error{
				Internal: trait.ErrApplicationStillUse,
				Detail:   edges,
			}
			ApplicationStillUseError.From(err.ToJson()).AbortGin(ctx)
			return
		}
	}
	jid, err := e.CreateDeleteJobAnStart(ctx, ains.AID, sid)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrComponentNotFound) {
			ComponentNotfoundError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrJobExecuting) {
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, app.ErrNoAvailableWorker) {
			err.Detail = fmt.Errorf("job %d has been create, but %s, please try exeute it later", jid, err.Error())
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrConfigValidate) {
			ParamError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, jid)
	}
}

func hasOuterEdge(edges map[int][]int) [][2]int {
	outer := [][2]int{}
	for to, froms := range edges {
		for _, from := range froms {
			if _, ok := edges[from]; !ok {
				outer = append(outer, [2]int{from, to})
			}
		}
	}
	return outer
}

// DeleteJobAndStart delete job target application
//
//	@Summary		将对应任务更改为删除类型并执行
//	@Description	将对应任务更改为删除类型并执行
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			jid			path		int			true	"任务ID"
//
//	@Success		200			{object}	int			"返回为空"
//	@Failure		400			{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404			{object}	HTTPError	"对象不存在，不允许更新"
//	@Failure		412			{object}	HTTPError	"不允许设置任务配置"
//	@Failure		500			{object}	HTTPError	"系统内部错误"
//	@Router			/job/{jid} 	[DELETE]
func (e *ExecutorEngine) MarkDeleteJobAndStart(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, err0 := strconv.Atoi(jidStr)
	if err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}

	err := e.Store.MarkJobDelete(ctx, jid)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrJobExecuting) {
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrConfigValidate) {
			ParamError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	}

	err = e.Executor.StartJob(ctx, jid)
	if err != nil {
		if trait.IsInternalError(err, app.ErrNoAvailableWorker) {
			err.Detail = fmt.Errorf("job %d has been create, but %s, please try exeute it later", jid, err.Error())
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, nil)
	}
}

// GetSnapshotJobSchemaWithName get the  job with application schema
//
//	@Summary		获取以应用包为目标版本，以当前系统运行应用实例为配置的，即将需要执行的任务配置信息
//	@Description	获取以应用包为目标版本，以当前系统运行应用实例为配置的，即将需要执行的任务配置信息，但不会生成真正的job，需要配合接口使用
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			tid												query		[]int		false	"应用配置模板ID,优先级按输入顺序,自低向高。如'http://127.0.0.1/test?tid=1&tid=2'"
//	@Param			sid												query		int			false	"系统空间ID,多实例下必填"
//	@Param			name											path		string		true	"应用包名"
//	@Param			version											path		string		true	"应用包版本"
//	@Success		200												{object}	jobSchema	"任务jsonschema数据"
//	@Failure		400												{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404												{object}	HTTPError	"对象不存在"
//	@Failure		500												{object}	HTTPError	"系统内部错误"
//	@Router			/job/jsonschema/snapshot/name/{name}/{version} 	[get]
func (e *ExecutorEngine) GetSnapshotJobSchemaWithName(ctx *gin.Context) {
	a := ctx.Param("name")
	v := ctx.Param("version")
	if v == "" {
		sid := e.getSidFromContext(ctx)
		if ctx.IsAborted() {
			return
		}
		ains, err := e.GetWorkAPPIns(ctx, a, sid)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From("no work application instance in the system:" + err.Error()).AbortGin(ctx)
			return
		} else if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		v = ains.Version
	}
	aid, err := e.GetAPPID(ctx, a, v)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
	} else if err == nil {
		e.getSnapshotJobSchema(ctx, aid)
	} else {
		UnknownError.From(err.Error()).AbortGin(ctx)
	}
}

// GetSnapshotJobSchema get the  job with application schema
//
//	@Summary		获取以应用包为目标版本，以当前系统运行应用实例为配置的，即将需要执行的任务配置信息
//	@Description	获取以应用包为目标版本，以当前系统运行应用实例为配置的，即将需要执行的任务配置信息，但不会生成真正的job，需要配合接口使用
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			tid								query		[]int		false	"应用配置模板ID,优先级按输入顺序,自低向高。如'http://127.0.0.1/test?tid=1&tid=2'"
//	@Param			sid								query		int			false	"系统空间ID,多实例下必填"
//	@Param			aid								path		int			true	"应用包ID"
//
//	@Success		200								{object}	jobSchema	"任务jsonschema数据"
//	@Failure		400								{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404								{object}	HTTPError	"对象不存在"
//	@Failure		500								{object}	HTTPError	"系统内部错误"
//	@Router			/job/jsonschema/snapshot/{aid} 	[get]
func (e *ExecutorEngine) GetSnapshotJobSchema(ctx *gin.Context) {
	aidStr := ctx.Param("aid")
	aid, err0 := strconv.Atoi(aidStr)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the aid is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}

	e.getSnapshotJobSchema(ctx, aid)
}

func (e *ExecutorEngine) getSidFromContext(ctx *gin.Context) int {
	sid := 0
	if e.SID >= 0 {
		sid = e.SID
	} else {
		queryInt, err := parseIntFromQuery(ctx, "sid")
		if err != nil {
			err.AbortGin(ctx)
			return -1
		}
		sid = queryInt[0]
	}
	return sid
}

func (e *ExecutorEngine) getSnapshotJobSchema(ctx *gin.Context, aid int) {
	sid := e.getSidFromContext(ctx)
	if ctx.IsAborted() {
		return
	}
	tids := ctx.QueryArray("tid")
	configs := make([]*trait.ApplicationInstance, 0, len(tids)+1)
	for _, tid := range tids {
		id, err0 := strconv.Atoi(tid)
		if err0 != nil {
			ParamError.From(fmt.Sprintf("the tid '%s' is not a int, error: %s", tid, err0.Error())).AbortGin(ctx)
			return
		}
		ct, err := e.Store.GetConfigTemplate(ctx, id)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			ParamError.From(fmt.Sprintf("the config template with tid '%d' not found, it may has been deleted, try other", id)).AbortGin(ctx)
			return
		}
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		ins := &trait.ApplicationInstance{
			AppConfig: ct.Config.AppConfig,
		}
		for cname, obj := range ct.Config.Components {
			obj.Component.Name = cname
			ins.Components = append(ins.Components, obj)
		}
		configs = append(configs, ins)
	}

	jb, err := e.Store.NewFakeJobRecord(ctx, aid, sid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if lang := ctx.Query("lang"); lang != "" {
		alias, err := e.GetAppLang(ctx, lang, jb.Target.AName, AppZone)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		jb.Target.Alias = alias
	}

	configs = append(configs, jb.Target)

	// merge config template for select
	e.Store.MergeJobConfigs(jb, configs...)

	app := jb.Target.Application
	if err := builder.SetAPPUISchema(&app); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	coms := make(map[string]*trait.ComponentInstance, len(jb.Target.Components))
	for _, c := range jb.Target.Components {
		coms[c.Component.Name] = c
	}
	js, err0 := convertJobIntoJobSchema(app, jb)
	if err0 != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
	}

	ctx.JSON(http.StatusOK, js)
}

func convertJobIntoJobSchema(app trait.Application, jb *trait.JobRecord) (jobSch jobSchema, err error) {
	coms := make(map[string]*trait.ComponentInstance, len(jb.Target.Components))
	for _, c := range jb.Target.Components {
		coms[c.Component.Name] = c
	}
	uisch, err0 := app.NewUISChema()
	if err0 != nil {
		err = err0
		return
	}
	jobSch = jobSchema{
		JID: jb.ID,
		ApplicationInstanceOverview: trait.ApplicationInstanceOverview{
			ApplicationMeta:         jb.Target.ApplicationMeta,
			ApplicationinstanceMeta: jb.Target.ApplicationinstanceMeta,
		},
		Schema:   newSchemaFromApplication(&app),
		UISchema: uisch,
		FromData: appInstanceConfig{
			AppConfig:  jb.Target.AppConfig,
			Components: coms,
		},
	}
	return
}

// GetJobschema get the  job with application schema
//
//	@Summary		获取任务详细信息与配置
//	@Description	获取指定任务的详细信息、配置与状态等
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			lang					query		string		false	"语言参数"
//	@Param			jid						path		int			true	"任务ID"
//
//	@Success		200						{object}	jobSchema	"任务jsonschema数据"
//	@Failure		400						{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404						{object}	HTTPError	"对象不存在"
//	@Failure		500						{object}	HTTPError	"系统内部错误"
//	@Router			/job/jsonschema/{jid} 	[get]
func (e *ExecutorEngine) GetJobschema(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, err0 := strconv.Atoi(jidStr)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the offset is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}
	tx, err := e.Store.Begin(ctx)
	if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			e.Log.Errorf("GetJobschema close transaction error: %s", err.Error())
		}
	}()
	jb, err := tx.GetJobRecord(ctx, jid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	if lang := ctx.Query("lang"); lang != "" {
		alias, err := e.GetAppLang(ctx, lang, jb.Target.AName, AppZone)
		if err != nil {
			UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		jb.Target.Alias = alias
	}
	app, err := tx.GetAPP(ctx, jb.Target.AID)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	if err := builder.SetAPPUISchema(app); err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	js, err0 := convertJobIntoJobSchema(*app, &jb)
	if err0 != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}

	ctx.JSON(http.StatusOK, js)
}

type jobSchema struct {
	// job id
	trait.ApplicationInstanceOverview `json:",inline"`
	JID                               int `json:"jid"`
	// 应用实例配置数据,使用json反序列化为对象后，配合schema使用
	FromData appInstanceConfig `json:"formData"`
	// 应用包整体配置文档，使用json反序列化为对象后使用
	Schema   appSChema                 `json:"schema"`
	UISchema trait.ApplicationUISchema `json:"uiSchema,omitempty"`
}

type (
	JobShcema         = jobSChemaConfig
	AppInstanceConfig = appInstanceConfig
)

type jobSChemaConfig struct {
	trait.ApplicationInstanceOverview `json:",inline"`
	// 应用实例配置数据,使用json反序列化为对象后，配合schema使用
	FromData appInstanceConfig `json:"formData"`
}

func (c *jobSChemaConfig) ToAppIns() *trait.ApplicationInstance {
	coms := make([]*trait.ComponentInstance, 0, len(c.FromData.Components))
	for name, c := range c.FromData.Components {
		c.Component.Name = name
		coms = append(coms, c)
	}
	return &trait.ApplicationInstance{
		Application: trait.Application{
			ApplicationMeta: c.ApplicationMeta,
		},
		ApplicationinstanceMeta: trait.ApplicationinstanceMeta{
			Comment: c.Comment,
			OType:   c.OType,
		},
		AppConfig:  c.FromData.AppConfig,
		Components: coms,
		Trait:      c.FromData.Trait,
	}
}

type appInstanceConfig struct {
	Trait trait.ApplicationTrait `json:"trait,omitempty"`
	// 应用级配置
	AppConfig map[string]interface{} `json:"appConfig,omitempty"`
	// 各个组件配置
	Components map[string]*trait.ComponentInstance `json:"components"`
}

// GetJob get the application's job
//
//	@Summary		获取任务详细信息与配置
//	@Description	获取指定任务的详细信息、配置与状态等
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			lang		query		string			false	"语言参数"
//	@Param			jid			path		int				true	"任务ID"
//
//	@Success		200			{object}	trait.JobRecord	"任务ID"
//	@Failure		400			{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404			{object}	HTTPError		"对象不存在"
//	@Failure		500			{object}	HTTPError		"系统内部错误"
//	@Router			/job/{jid} 	[get]
func (e *ExecutorEngine) GetJob(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, err0 := strconv.Atoi(jidStr)
	if err0 != nil {
		ParamError.From(fmt.Sprintf("the offset is not a int, error: %s", err0.Error())).AbortGin(ctx)
		return
	}
	jb, err := e.Store.GetJobRecord(ctx, jid)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		if lang := ctx.Query("lang"); lang != "" {
			alias, err := e.GetAppLang(ctx, lang, jb.Target.AName, AppZone)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			jb.Target.Alias = alias
		}
		ctx.JSON(http.StatusOK, jb)
	}
}

// ListJob list job
//
//	@Summary		获取任务列表
//	@Description	获取任务列表,列表内为概要信息,支持分页滚动查询
//	@Description	列表内为概要信息,只有当前应用实例ID，目标应用实例ID、状态名称和任务ID为有效数据
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			offset	query		int			false	"分页偏移量"
//	@Param			sid		query		int			false	"系统ID"
//	@Param			name	query		string		false	"应用名称,区分大小写"
//	@Param			status	query		[]string	false	"状态过滤器设置,多个状态间关系为'或关系'"
//	@Param			limit	query		int			false	"分页大小"
//	@Param			jtype	query		[]int		false	"任务类型过滤,默认不过滤"
//	@Param			title	query		string		false	"对应语言的包名"
//	@Param			lang	query		string		false	"语言参数"
//
//	@Success		200		{object}	pageJob		"任务列表"
//	@Failure		400		{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404		{object}	HTTPError	"对象不存在"
//	@Failure		500		{object}	HTTPError	"系统内部错误"
//	@Router			/job 	[get]
func (e *ExecutorEngine) ListJob(ctx *gin.Context) {
	name := ctx.Query("name")
	status, err1 := ConvertStringToIntArray(ctx.QueryArray("status")...)
	if err1 != nil {
		ParamError.From(fmt.Sprintf("status must set with int array, convert error: %s", err1.Error())).AbortGin(ctx)
		return
	}
	jtypes, err1 := ConvertStringToIntArray(ctx.QueryArray("jtype")...)
	if err1 != nil {
		ParamError.From(fmt.Sprintf("status must set with int array, convert error: %s", err1.Error())).AbortGin(ctx)
		return
	}
	query, err := ParseIntFromQueryWithDefault(
		ctx,
		[]string{"limit", "offset", "sid"},
		"20", "0", "-1")
	if err != nil {
		err.AbortGin(ctx)
		return
	}
	sid := query[2]
	if e.SID >= 0 {
		sid = e.SID
	}

	if name == "" {
		lang := ctx.Query("lang")
		alias := ctx.Query("title")
		aname := e.Store.GetAname(lang, alias, AppZone)
		if aname == "" {
			aname = alias
		}
		name = aname
	}

	filter := &trait.AppInsFilter{
		Status: status,
		Name:   name,
		Sid:    sid,
		Limit:  query[0],
		Offset: query[1],
		Jtype:  jtypes,
	}
	count, err0 := e.Store.CountJobRecord(ctx, filter)
	if err0 != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
		return
	}

	jb, err0 := e.Store.ListJobRecord(ctx, filter)
	if err0 != nil {
		UnknownError.From(err0.Error()).AbortGin(ctx)
	} else {
		if lang := ctx.Query("lang"); lang != "" {
			for _, j := range jb {
				alias, err := e.GetAppLang(ctx, lang, j.Target.AName, AppZone)
				if err != nil {
					UnknownError.From(err.Error()).AbortGin(ctx)
					return
				}
				j.Target.Alias = alias
			}
		}

		ctx.JSON(http.StatusOK, pageJob{
			TotalNum: count,
			Data:     jb,
		})
	}
}

// ListJobLog list job log
//
//	@Summary		获取任务简易过程日志
//	@Description	获取任务简易过程日志, 日志为任务执行过程中部分执行操作日志,有些操作日志存在截断。
//	@Description	该接口主要用于从任务视图或组件视图视角查看任务执行信息
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			offset		query		int			false	"分页偏移量"
//	@Param			limit		query		int			false	"分页大小"
//	@Param			jid			query		int			false	"设定任务ID过滤条件,常用于任务视图,-1或不设置时为不过滤"
//	@Param			cid			query		int			false	"设定组件实例ID过滤条件,不设置jid时为组件视图,也可以结合jid组合过滤，-1或不设置时为不过滤"
//	@Param			timestamp	query		int			false	"秒级时间戳过滤条件，负数为时间戳以前，正数为时间戳以后，0或不设置为不过滤"
//	@Param			sort		query		string		false	"排序方式,默认降序"
//	@Param			count		query		string		false	"是否返回对应过滤条件下的数据数量，false或False为否，其余为是，一般在分页查询仅第一次查询时设置为是"
//	@Param			lang		query		string		false	"语言参数"
//
//	@Success		200			{object}	JobLogList	"任务列表"
//	@Failure		400			{object}	HTTPError	"客户端请求参数错误"
//	@Failure		404			{object}	HTTPError	"对象不存在"
//	@Failure		500			{object}	HTTPError	"系统内部错误"
//	@Router			/job/log 	[get]
func (e *ExecutorEngine) ListJobLog(ctx *gin.Context) {
	offset := ctx.DefaultQuery("offset", "0")
	limit := ctx.DefaultQuery("limit", "10")
	jid := ctx.DefaultQuery("jid", "-1")
	cid := ctx.DefaultQuery("cid", "-1")
	timestamp := ctx.DefaultQuery("timestamp", "0")
	sortType := ctx.DefaultQuery("sort", string(trait.DescSortType))
	needCount := ctx.DefaultQuery("count", "True")

	params, err0 := ConvertStringToIntArray(offset, limit, jid, cid, timestamp)
	if err0 != nil {
		ParamError.From(fmt.Sprintf(`query parama "offset", "limit", "jid", "cid" must is nil or interger , convert error: %s`, err0.Error())).AbortGin(ctx)
		return
	}
	f := trait.JobLogFilter{
		Offset:   params[0],
		Limit:    params[1],
		JID:      params[2],
		CID:      params[3],
		Timestmp: params[4],
		Sort:     trait.SortType(sortType),
	}

	res := &JobLogList{}
	if needCount != "False" && needCount != "false" {
		f.Offset = 0
		count, err1 := e.CountJobLog(ctx, f)
		if err1 != nil {
			UnknownError.From(err1.Error()).AbortGin(ctx)
			return
		}
		res.TotalNum = count
	}
	f.Offset = params[0]

	jls, err1 := e.Store.ListJobLog(ctx, f)
	if err1 != nil {
		UnknownError.From(err1.Error()).AbortGin(ctx)
		return
	}
	data := make([]JobLog, len(jls))
	lang := ctx.Query("lang")

	for i, l := range jls {
		alias := l.Aname
		if lang != "" {
			a, err := e.GetAppLang(ctx, lang, l.Aname, AppZone)
			if err != nil {
				UnknownError.From(err.Error()).AbortGin(ctx)
				return
			}
			alias = a
		}
		l.Alias = alias

		data[i].JobLog = l
		data[i].Description = codes.ErrorCache.GetCode(l.Code).Description
		data[i].Msg = string(l.Msg)
	}
	res.Data = data
	ctx.JSON(http.StatusOK, res)
}

type JobLogList struct {
	TotalNum int      `json:"totalNum"`
	Data     []JobLog `json:"data"`
}

type JobLog struct {
	trait.JobLog `json:",inline"`
	// 错误码对应的简述
	Description string `json:"description"`
}

type pageJob struct {
	TotalNum int               `json:"totalNum"`
	Data     []trait.JobRecord `json:"data"`
}

// SetJobConfig set job target application instance config
//
//	@Summary		配置任务中目标应用
//	@Description	配置任务中目标应用配置,用于后续安装/升级任务中将实际应用配置为期望配置
//	@Description	每次变更配置后不会立即生效,需要重新执行该任务
//
//	@Tags			job
//	@Accept			json
//	@Produce		json
//
//	@Param			config		body		jobSChemaConfig	true	"目标应用配置"
//	@Param			jid			path		int				true	"任务ID"
//
//	@Success		200			{object}	int				"返回为空"
//	@Failure		400			{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404			{object}	HTTPError		"对象不存在，不允许更新"
//	@Failure		412			{object}	HTTPError		"不允许设置任务配置"
//	@Failure		500			{object}	HTTPError		"系统内部错误"
//	@Router			/job/{jid} 	[PUT]
func (e *ExecutorEngine) SetJobConfig(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, err0 := strconv.Atoi(jidStr)
	if err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	a := &jobSChemaConfig{}
	if err0 := ctx.BindJSON(a); err0 != nil {
		ParamError.From(err0.Error()).AbortGin(ctx)
		return
	}
	ains := a.ToAppIns()
	err := e.Store.SetJobConfig(ctx, jid, ains)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			NotFoundError.From(err.Error()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrJobExecuting) {
			ConditionError.From(err.ToJson()).AbortGin(ctx)
		} else if trait.IsInternalError(err, trait.ErrConfigValidate) {
			ParamError.From(err.Error()).AbortGin(ctx)
		} else {
			UnknownError.From(err.Error()).AbortGin(ctx)
		}
	} else {
		ctx.JSON(http.StatusOK, nil)
	}
}

// GetComponentInstance get component instance detail
//
//	@Summary		获取组件实例信息与配置
//	@Description	获取指定组件实例信息与配置
//
//	@Tags			componentInstance
//	@Accept			json
//	@Produce		json
//
//	@Param			cid							path		int						true	"组件实例ID"
//
//	@Success		200							{object}	componentInstanceConfig	"任务jsonschema数据"
//	@Failure		400							{object}	HTTPError				"客户端请求参数错误"
//	@Failure		404							{object}	HTTPError				"对象不存在"
//	@Failure		500							{object}	HTTPError				"系统内部错误"
//	@Router			/component/instance/{cid} 	[get]
func (e *ExecutorEngine) GetComponentInstance(ctx *gin.Context) {
	cid, err0 := ConvertStringToIntArray(ctx.Param("cid"))
	if err0 != nil {
		ParamError.From(fmt.Sprintf("cid must set a integer, convert error: %s", err0.Error())).AbortGin(ctx)
		return
	}
	cins, err := e.GetComponentIns(ctx, cid[0])
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	com, err := e.GetAPPComponent(ctx, cins.Acid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		NotFoundError.From(fmt.Sprintf("the component meta not found, error: %s", err.Error())).AbortGin(ctx)
		return
	} else if err != nil {
		UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, componentInstanceConfig{
		FromData: *cins,
		Schema:   newComponentSchema(com),
	})
}

type componentInstanceConfig struct {
	// 应用实例配置数据,使用json反序列化为对象后，配合schema使用
	FromData trait.ComponentInstance `json:"formData"`
	// 应用包整体配置文档，使用json反序列化为对象后使用
	Schema componentSchema `json:"schema"`
}
