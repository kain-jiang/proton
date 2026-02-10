package compose

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"taskrunner/api/rest"
	"taskrunner/api/rest/proton_component"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type HTTPError = rest.HTTPError

type Server struct {
	trait.ComposeJobWriter
	protonComponentOperator
	*applicationOperator
	Log  *logrus.Logger
	lock *sync.Mutex
	jobs sync.Map
	ctx  context.Context
}

func NewServer(ctx context.Context, store trait.ComposeJobWriter, s proton_component.GinServer, engine *rest.ExecutorEngine, kcli kubernetes.Interface, log *logrus.Logger) *Server {
	return &Server{
		ComposeJobWriter:        store,
		protonComponentOperator: newProtonComponentOPerator(s, log),
		applicationOperator:     newApplicationOperator(engine, kcli),
		Log:                     log,
		lock:                    &sync.Mutex{},
		jobs:                    sync.Map{},
		ctx:                     ctx,
	}
}

type composeJob struct {
	job    *trait.ComposeJob
	ctx    context.Context
	cancel func(*trait.Error)
	lock   *sync.Mutex
}

func newComposeJob(job *trait.ComposeJob, ctx context.Context) *composeJob {
	jctx, cancel := trait.WithCancelCauesContext(ctx)
	j := &composeJob{
		job:    job,
		ctx:    jctx,
		cancel: cancel,
		lock:   &sync.Mutex{},
	}
	return j
}

func (s *Server) RegistryHandler(r *gin.RouterGroup) {
	r.POST("/composejob", s.Create)
	r.GET("/composejob", s.ListComposeJobs)
	r.POST("/composejob/:jid", s.Execute)
	r.PATCH("/composejob/:jid", s.Patch)
	r.PUT("/composejob/:jid", s.TryStop)
	r.GET("/composejob/:jid", s.GetJob)
	r.GET("/manifests/:name/:version", s.GetManifests)
	r.GET("/manifests/work", s.ListWorkManifests)
	r.GET("/manifests", s.ListManifests)
	r.POST("/manifests", s.UploadManifests)
}

func (s *Server) Recovery(ctx context.Context) *trait.Error {
	offset := 0
	filter := trait.ComposeJobFilter{
		Status: []int{trait.AppDoingStatus, trait.AppStopingStatus},
	}
	for {
		jobs, _, err := s.ListComposeJob(ctx, 100, offset, filter)
		if err != nil {
			return err
		}
		if len(jobs) == 0 {
			return nil
		}
		offset += len(jobs)
		s.log.Tracef("compose job recover offset: %d", offset)
		for _, job := range jobs {
			if err := s.ReStartComposeJob(ctx, job.Jid); err != nil {
				if !trait.IsInternalError(err, trait.ErrNotFound) {
					s.log.Errorf("recover compose job %d error: %s", job.Jid, err.Error())
					return err
				}
			}
			if job.Status == trait.AppStopingStatus {
				if err := s.StopComposeJob(ctx, job.Jid); err != nil {
					if trait.IsInternalError(err, trait.ErrNotFound) {
						continue
					}
					return err
				}
			}
		}
	}
}

// Patch 修改任务配置
//
//	@Summary		修改批量执行任务配置
//	@Description	修改批量执行任务配置,已安装完成配置修改不会生效
//
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//
//	@Param			config		body		ManifestsJob	true	"套件配置"
//	@Param			jid			path		int				true	"任务ID"
//	@Param			app-format	header		string			false	"header值为schema时请求体格式为套件配置格式"
//
//	@Success		200			{object}	int				"返回为空"
//	@Failure		400			{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404			{object}	HTTPError		"对象不存在，不允许更新"
//	@Failure		412			{object}	HTTPError		"不允许设置任务配置"
//	@Failure		500			{object}	HTTPError		"系统内部错误"
//	@Router			/composejob/{jid} [post]
func (s *Server) Patch(ctx *gin.Context) {
	jidstr := ctx.Param("jid")
	jid, rerr := strconv.Atoi(jidstr)
	if rerr != nil {
		rest.ParamError.From(
			fmt.Sprintf("jid must is int, now: %s, err: %s",
				jidstr, rerr.Error())).
			AbortGin(ctx)
		return
	}
	s.lock.Lock()
	_, ok := s.jobs.Load(jid)
	if ok {
		s.lock.Unlock()
		rest.ConditionError.From("job can't path when running, try to stop it before patch").AbortGin(ctx)
		return
	}
	s.jobs.Store(jid, nil)
	s.lock.Unlock()
	defer s.jobs.Delete(jid)

	obj := trait.ComposeJob{}
	format := ctx.GetHeader("app-format")
	if format == "schema" {
		mj := &ManifestsJob{}
		if rerr := ctx.BindJSON(mj); rerr != nil {
			rest.ParamError.From(rerr.Error()).AbortGin(ctx)
			return
		}
		obj = *mj.ComposeJob()
	} else {
		if rerr := ctx.BindJSON(&obj); rerr != nil {
			rest.ParamError.From(rerr.Error()).AbortGin(ctx)
			return
		}
	}

	cur, err := s.GetComposeJob(ctx, jid)
	if err != nil {
		rest.NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	}

	// switch cur.Status {
	// case trait.AppDoingStatus:
	// 	rest.ConditionError.From("job can't path when running, try to stop it before patch").AbortGin(ctx)
	// 	return
	// }

	if err := s.PatchComposeJob(ctx, obj, *cur); err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, nil)
}

func (s *Server) PatchComposeJob(ctx context.Context, path trait.ComposeJob, cur trait.ComposeJob) *trait.Error {
	i := 0
	for j := range cur.Config.ProtonComponent {
		if i < cur.Processed {
			i++
			continue
		}
		i++
		cur.Config.ProtonComponent[j] = path.Config.ProtonComponent[j]
	}

	tx, err := s.Begin(ctx)
	if err != nil {
		return err
	}
	for j := range cur.Config.AppConfig {
		if i < cur.Processed {
			i++
			continue
		}
		i++
		ajid := -1
		if err = doUntilSuccess(func() *trait.Error {
			jid, gerr := tx.GetCompoesJobTask(ctx, cur.Jid, i)
			if trait.IsInternalError(gerr, trait.ErrNotFound) {
				gerr = nil
			}
			if gerr != nil {
				return gerr
			}
			ajid = jid
			return nil
		}, s.Log); err != nil {
			s.Log.Errorf("get compose job's application instance task error: %s", err.Error())
			goto finish
		}

		if ajid >= 0 {
			_, ierr := tx.GetAPPIns(ctx, ajid)
			if ierr != nil {
				err = ierr
				s.Log.Errorf("get compose job's application instance detail error: %s", err.Error())
				goto finish
			}
			if ierr := tx.UpdateAPPInsConfig(ctx, *path.Config.AppConfig[j]); ierr != nil {
				err = ierr
				s.Log.Errorf("update  compose job's application instance config error: %s", err.Error())
				goto finish
			}
		}
		cur.Config.AppConfig[j] = path.Config.AppConfig[j]
	}

	err = tx.SetComposeJob(ctx, cur)
	if err != nil {
		s.Log.Errorf("update  compose job's config error: %s", err.Error())
	}

finish:
	if err != nil {
		if ierr := tx.Rollback(); ierr != nil {
			s.Log.Errorf("update  compose job's config fail and rollback return error: %s", err.Error())
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		s.Log.Errorf("update compose job's config transaction commit error: %s", err)
	}
	return err
}

// Create 创建批量执行任务
//
//	@Summary		创建批量执行任务
//	@Description	创建批量执行任务,当设置header字段app-formatt为schema值时接收套件配置格式
//
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//
//	@Param			config		body		ManifestsJob	true	"套件配置"
//	@Param			app-format	header		string			false	"header值为schema时请求体格式为套件配置格式"
//
//	@Success		200			{object}	int				"返回为任务ID"
//	@Failure		400			{object}	HTTPError		"客户端请求参数错误"
//	@Failure		404			{object}	HTTPError		"对象不存在，不允许更新"
//	@Failure		412			{object}	HTTPError		"不允许设置任务配置"
//	@Failure		500			{object}	HTTPError		"系统内部错误"
//	@Router			/composejob [post]
func (s *Server) Create(ctx *gin.Context) {
	obj := trait.ComposeJob{}

	format := ctx.GetHeader("app-format")
	if format == "schema" {
		mj := &ManifestsJob{}
		if rerr := ctx.BindJSON(mj); rerr != nil {
			rest.ParamError.From(rerr.Error()).AbortGin(ctx)
			return
		}
		obj = *mj.ComposeJob()
	} else {
		if rerr := ctx.BindJSON(&obj); rerr != nil {
			rest.ParamError.From(rerr.Error()).AbortGin(ctx)
			return
		}
	}

	jid, err := s.NewComposeJob(ctx, obj)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		rest.NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	}
	if trait.IsInternalError(err, trait.ErrComponentNotFound) {
		rest.ComponentNotfoundError.From(err.ToJson()).AbortGin(ctx)
		return
	}
	if trait.IsInternalError(err, trait.ErrApplicationNotFound) {
		rest.ApplicationNotfoundError.From(err.ToJson()).AbortGin(ctx)
		return
	}
	if err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, jid)
}

// Execute try stop job
//
//	@Summary		启动任务
//	@Description	通过任务ID尝试启动任务，应通过查询接口查看最终状态
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	int			"返回空"
//	@Failure		400	{object}	HTTPError	"参数错误"
//	@Failure		404	{object}	HTTPError	"对象不存在"
//	@Failure		500	{object}	HTTPError	"系统异常或内部错误"
//	@Router			/composejob/{jid} [POST]
func (s *Server) Execute(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, rerr := strconv.Atoi(jidStr)
	if rerr != nil {
		rest.ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	err := s.ReStartComposeJob(ctx, jid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		rest.NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	}
	if err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, jid)
}

// TryStop try stop job
//
//	@Summary		尝试暂停任务
//	@Description	通过任务ID尝试暂停任务，应通过查询接口查看最终状态
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	int			"返回空"
//	@Failure		400	{object}	HTTPError	"参数错误"
//	@Failure		404	{object}	HTTPError	"对象不存在"
//	@Failure		500	{object}	HTTPError	"系统异常或内部错误"
//	@Router			/composejob/{jid} [PUT]
func (s *Server) TryStop(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, rerr := strconv.Atoi(jidStr)
	if rerr != nil {
		rest.ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	err := s.StopComposeJob(ctx, jid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		rest.NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	}
	if err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	ctx.JSON(http.StatusOK, jid)
}

func (s *Server) Job2Schema(ctx *gin.Context, job *trait.ComposeJob) {
	ids, err := s.GetCompoesJobTasks(ctx, job.Jid)
	if err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i][0] < ids[j][0]
	})
	for _, i := range job.Config.AppConfig {
		i.Status = trait.AppToCreateStatus
	}
	for _, i := range ids {
		job.Config.AppConfig[i[0]].ID = i[1]
		job.Config.AppConfig[i[0]].Status = trait.AppSucessStatus
	}
	last := len(ids) - 1
	if last >= 0 && job.Processed != job.Total {
		si := ids[last][0]
		ains, err := s.GetAPPIns(ctx, ids[last][1])
		if trait.IsInternalError(err, trait.ErrNotFound) {
			ains = &trait.ApplicationInstance{}
			ains.Status = trait.AppFailStatus
		} else if err != nil {
			rest.UnknownError.From(fmt.Sprintf("get sub application task error: %s", err.Error())).AbortGin(ctx)
			return
		}
		job.Config.AppConfig[si].Status = ains.Status
	}
	if lang := ctx.Query("lang"); lang != "" {
		err := trait.ConvertLangs(ctx, s, job.Config.AppConfig, lang, rest.AppZone)
		if err != nil {
			rest.UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		alias, err := s.GetAppLang(ctx, lang, job.Jname, manifestZone)
		if err != nil {
			rest.UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
		job.Title = alias
	}
	ctx.JSON(http.StatusOK, ComposeJob2ManifestsJob(job))
}

// GetJob get compose job detail from store
//
//	@Summary		获取任务信息
//	@Description	通过任务ID获取任务信息
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//	@Param			app-format	header		string			false	"header值为schema时相应格式为套件配置格式"
//	@Success		200			{object}	ManifestsJob	"返回任务信息"
//	@Failure		400			{object}	HTTPError		"参数错误"
//	@Failure		404			{object}	HTTPError		"对象不存在"
//	@Failure		500			{object}	HTTPError		"系统异常或内部错误"
//	@Router			/composejob/{jid} [get]
func (s *Server) GetJob(ctx *gin.Context) {
	jidStr := ctx.Param("jid")
	jid, rerr := strconv.Atoi(jidStr)
	if rerr != nil {
		rest.ParamError.From(rerr.Error()).AbortGin(ctx)
		return
	}
	job, err := s.GetComposeJob(ctx, jid)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		rest.NotFoundError.From(err.Error()).AbortGin(ctx)
		return
	}
	if err != nil {
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}
	format := ctx.GetHeader("app-format")
	if format == "schema" {
		s.Job2Schema(ctx, job)
	} else {
		ctx.JSON(http.StatusOK, job)
	}
}

// ListComposeJobs get composejob detail from store
//
//	@Summary		批量操作任务列表
//	@Description	批量操作任务列表
//	@Tags			composejob
//	@Accept			json
//	@Produce		json
//	@Param			app-format	header		string				false	"header值为schema时相应格式为套件配置格式"
//	@Param			lang		query		string				false	"语言参数"
//	@Param			name		query		string				true	"套件名"
//	@Param			status		query		[]int				false	"状态过滤器设置,多个状态间关系为'或关系'"
//	@Param			title		query		string				false	"对应语言的包名"
//	@param			limit		query		int					false	"分页数"
//	@param			offset		query		int					false	"偏移量"
//	@param			type		query		int					false	"任务类型过滤,0为所有任务,1为一般任务,2为套件任务,默认为1"
//	@param			sid			query		int					false	"系统空间ID"
//
//	@Success		200			{object}	ComposeJobMetaList	"返回套件配置清单"
//	@Failure		400			{object}	HTTPError			"参数错误"
//	@Failure		404			{object}	HTTPError			"对象不存在"
//	@Failure		500			{object}	HTTPError			"系统异常或内部错误"
//	@Router			/composejob [get]
func (s *Server) ListComposeJobs(ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		lang := ctx.Query("lang")
		alias := ctx.Query("title")
		aname := s.Store.GetAname(lang, alias, manifestZone)
		if aname == "" {
			aname = alias
		}
		name = aname
	}
	status, err1 := rest.ConvertStringToIntArray(ctx.QueryArray("status")...)
	if err1 != nil {
		rest.ParamError.From(fmt.Sprintf("status must set with int array, convert error: %s", err1.Error())).AbortGin(ctx)
		return
	}

	querys, herr := rest.ParseIntFromQueryWithDefault(
		ctx,
		[]string{"limit", "offset", "sid", "type"},
		"10", "0", "-1", "1")
	if herr != nil {
		herr.AbortGin(ctx)
		return
	}
	f := trait.ComposeJobFilter{
		Name:   name,
		Status: status,
		SID:    querys[2],
	}

	f.ListType = querys[3]

	objs, total, err := s.ListComposeJob(ctx, querys[0], querys[1], f)
	if err != nil {
		s.Log.WithError(err).Error("ListComposeJob fail")
		rest.UnknownError.From(err.Error()).AbortGin(ctx)
		return
	}

	if lang := ctx.Query("lang"); lang != "" {
		err := trait.ConvertLangs(ctx, s, objs, lang, manifestZone)
		if err != nil {
			s.Log.WithError(err).Error("ConvertLangs fail")
			rest.UnknownError.From(err.Error()).AbortGin(ctx)
			return
		}
	}

	ctx.JSON(http.StatusOK, ComposeJobMetaList{
		Data:  objs,
		Total: total,
	})
}

func (s *Server) ReStartComposeJob(ctx context.Context, jid int) *trait.Error {
	job, err := s.GetComposeJob(ctx, jid)
	if err != nil {
		return err
	}
	s.lock.Lock()
	defer s.lock.Unlock()

	_, ok := s.jobs.Load(jid)
	if ok {
		// already start
		return nil
	}
	j := newComposeJob(job, s.ctx)
	s.jobs.Store(jid, j)
	go func() {
		if err := s.ExecuteComposeJob(j.ctx, *job); err != nil {
			s.Log.Errorf("exeute compose job %d, name: %s error: %s", job.Jid, job.Jname, err.Error())
		}
	}()
	return nil
}

func (s *Server) StopComposeJob(ctx context.Context, jid int) *trait.Error {
	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok := s.jobs.Load(jid)
	if !ok {
		endtime := int(time.Now().Unix())
		return s.UpdateComposeJobStatus(context.Background(), jid, trait.AppStopedStatus, -2, endtime)
	}
	if v == nil {
		// the job is patch status
		return &trait.Error{
			Internal: trait.ErrJobCantStop,
			Detail:   "the job can't stop now",
		}
	}
	job := v.(*composeJob)
	job.cancel(&trait.Error{
		Internal: trait.ECJobCancel,
	})
	s.Log.Tracef("receive stop compose job %d signal, try stop", jid)
	return nil
}

func (s *Server) NewComposeJob(ctx context.Context, job trait.ComposeJob) (int, *trait.Error) {
	sid, err := s.applicationOperator.CreateSystem(ctx, job.System)
	if err != nil {
		return -1, err
	}

	// TODO: 检查应用依赖，排序应用列表，检查组件依赖
	appDepMetas := make([]trait.AppDepMeta, 0, 0)
	appDepMetaMap := make(map[string][]trait.AppDepMeta, len(job.Config.AppConfig))
	appConfigMap := make(map[string]*trait.ApplicationInstance, len(job.Config.AppConfig))
	sortedAppConfig := make([]*trait.ApplicationInstance, 0, len(job.Config.AppConfig))
	{
		for idx, ac := range job.Config.AppConfig {
			a, err := s.ComposeJobWriter.GetAPP(ctx, ac.Application.AID)
			if err != nil {
				return -1, err
			}
			job.Config.AppConfig[idx].Application = *a // 还原完整 Application 似乎无效
			appDepMetas = append(appDepMetas, a.Dependence...)
			appDepMetaMap[ac.Application.AName] = a.Dependence
			appConfigMap[ac.Application.AName] = ac
		}

		appDepInThisComposeJob := func(_m trait.AppDepMeta) bool {
			for _, app := range job.Config.AppConfig {
				if app.Application.AName == _m.AName {
					return true
				}
			}
			return false
			// TODO: 不检测版本
		}

		// 1.开始检查应用依赖
		for _, appDepMeta := range appDepMetas {
			if appDepInThisComposeJob(appDepMeta) {
				// 依赖存在会被安装，则跳过
				continue
			}
			_, err := s.ComposeJobWriter.GetWorkAPPIns(ctx, appDepMeta.AName, sid)
			if err != nil {
				// 查询不到现有安装应用，直接返回错误
				// 包括 APPIns 不存在，查询 APPIns失败
				// if trait.IsInternalError(err, trait.ErrNotFound)
				if trait.IsInternalError(err, trait.ErrNotFound) {
					err = &trait.Error{
						Internal: trait.ErrApplicationNotFound,
						Detail:   appDepMeta.AName,
						Err:      fmt.Errorf("get application [%s] in system [%d] error:%s", appDepMeta.AName, sid, err.Error()),
					}
					return -1, err
				}
				return -1, err
			}
		}

		// 2.排序安装顺序
		tasks := make(map[string][]string, len(appDepMetaMap))
		for appName, appDepMeta := range appDepMetaMap {
			edges := make([]string, len(appDepMeta))
			for _, edge := range appDepMeta {
				edges = append(edges, edge.AName)
			}
			tasks[appName] = edges
		}
		sorted, err := utils.TopoSort(tasks, s.Log)
		if err != nil {
			return -1, &trait.Error{
				Err: err,
			}
		}
		s.Log.Debugf("sorted app list: %v", sorted)
		for _, appName := range sorted {
			sortedAppConfig = append(sortedAppConfig, appConfigMap[appName])
		}
		job.Config.AppConfig = sortedAppConfig

		// 3.检查组件依赖
		willExistComponents := make([]*trait.ComponentMeta, 0)
		willExistGraph := make([]trait.Edge, 0)
		for _, ac := range job.Config.AppConfig {
			a := ac.Application
			willExistComponents = append(willExistComponents, a.Component...)
			willExistGraph = append(willExistGraph, a.Graph...)
			aName := a.AName
			aVersion := a.Version

			err := s.CheckDepComponents(ctx, &trait.Application{
				Component: willExistComponents,
				Graph:     willExistGraph,
				ApplicationMeta: trait.ApplicationMeta{
					AName:   aName,
					Version: aVersion,
				},
			}, sid)
			if err != nil {
				return -1, err
			}
			s.Log.Debugf("check app depComponents passed: %s(%s)", aName, aVersion)
		}

	}

	// if err := s.applicationOperator.CreateAccessInfo(ctx, job.Config.AccessInfo, job.NameSpace); err != nil {
	// 	return -1, err
	// }
	job.System.SID = sid
	i := len(job.Config.ProtonComponent)
	i += len(job.Config.AppConfig)
	job.Total = i
	job.CreateTime = int(time.Now().Unix())
	job.StartTime = -1
	job.EndTime = -1
	job.Status = trait.AppDoingStatus
	jid, err := s.InsertComposeJob(ctx, job)
	if err != nil {
		return -1, err
	}
	job.Jid = jid

	j := newComposeJob(&job, s.ctx)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.jobs.Store(jid, j)
	go func() {
		_ = s.ExecuteComposeJob(j.ctx, job)
	}()
	return jid, nil
}

func doUntilSuccess(f func() *trait.Error, log *logrus.Logger) *trait.Error {
	for {
		err := f()
		if trait.IsInternalError(err, trait.ECJobCancel) || trait.IsInternalError(err, trait.ECContextEnd) || trait.IsInternalError(err, trait.ECExit) {
			return err
		}
		if err != nil {
			log.Error(err.Error())
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	return nil
}

func (s *Server) ExecuteComposeJob(ctx context.Context, job trait.ComposeJob) *trait.Error {
	i := 0
	var err *trait.Error
	defer func() {
		status := trait.AppSucessStatus
		if err != nil {
			s.Log.Errorf("compose job execute error: %s", err.Error())
			if trait.IsInternalError(err, trait.ECJobCancel) {
				status = trait.AppStopedStatus
			} else {
				status = trait.AppFailStatus
			}
		}
		_ = doUntilSuccess(
			func() *trait.Error {
				job.EndTime = int(time.Now().Unix())
				return s.UpdateComposeJobStatus(context.Background(), job.Jid, status, -2, job.EndTime)
			}, s.Log,
		)
		s.jobs.Delete(job.Jid)
	}()

	job.StartTime = int(time.Now().Unix())
	err = doUntilSuccess(
		func() *trait.Error {
			err = s.UpdateComposeJobStatus(ctx, job.Jid, trait.AppDoingStatus, job.StartTime, -1)
			if err != nil {
				return err
			}
			return nil
		}, s.Log,
	)

	//
	for _, c := range job.Config.ProtonComponent {
		if i != job.Processed {
			// `process`ed ignore
			i++
			continue
		}
		if err = s.InstallProtonComponent(ctx, c, job.ComposeJobMeata); err != nil {
			return err
		}
		i++
		job.Processed++
		if err = doUntilSuccess(func() *trait.Error {
			return s.UpdateComposeJobProcess(ctx, job.Jid, i)
		}, s.Log); err != nil {
			return err
		}
	}

	for _, a := range job.Config.AppConfig {
		if i != job.Processed {
			// processed ignore
			i++
			continue
		}

		j := i
		i++
		ajid := -1
		if err = doUntilSuccess(func() *trait.Error {
			jid, gerr := s.GetCompoesJobTask(ctx, job.Jid, j)
			if trait.IsInternalError(gerr, trait.ErrNotFound) {
				gerr = nil
			}
			if gerr != nil {
				return gerr
			}
			ajid = jid
			return nil
		}, s.Log); err != nil {
			return err
		}

		if ajid == -1 {
			// create job
			a.SID = job.SID
			ajid, err = s.applicationOperator.CreateJob(ctx, a)
			if err != nil {
				return err
			}
			// store jid
			if err = doUntilSuccess(func() *trait.Error {
				return s.SetComposeJobTask(ctx, job.Jid, j, ajid)
			}, s.Log); err != nil {
				return err
			}
		}

		if err = func() *trait.Error {
			// wait job
			if err = s.StartJob(ctx, ajid); err != nil {
				return err
			}

			defer func() {
				// if job cancel stop try stop application job
				if trait.IsInternalError(err, trait.ECJobCancel) {
					_ = s.StopAndWaitJob(context.Background(), ajid)
				}
			}()
			status, gerr := s.WaitJob(ctx, ajid)
			if gerr != nil || status != trait.AppSucessStatus {
				if gerr == nil {
					gerr = &trait.Error{
						Internal: trait.ECNULL,
						Err:      fmt.Errorf("application job execute fail,  receive status %d", status),
						Detail: fmt.Sprintf("compose job id:%d name: %s,execute application job name: %s, version: %s, job id: %d",
							job.Jid, job.Jname, a.AName, a.Version, ajid),
					}
				}
				err = gerr
				return gerr
			} else {
				if err = doUntilSuccess(func() *trait.Error {
					return s.UpdateComposeJobProcess(ctx, job.Jid, i)
				}, s.Log); err != nil {
					return err
				}
				job.Processed++
			}
			return nil
		}(); err != nil {
			return err
		}

	}

	// finish
	job.EndTime = int(time.Now().Unix())
	job.Status = trait.AppSucessStatus
	err = doUntilSuccess(func() *trait.Error {
		return s.UpdateComposeJobStatus(context.Background(), job.Jid, trait.AppSucessStatus, -2, job.EndTime)
	}, s.Log)

	if job.Mversion != "" {
		if err != nil {
			return err
		}
		err = doUntilSuccess(func() *trait.Error {
			return s.InsertWorkComposeManifests(ctx, job.ComposeJobMeata)
		}, s.Log)
	}

	return err
}
