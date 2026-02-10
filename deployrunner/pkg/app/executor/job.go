package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component"
	"taskrunner/pkg/graph"
	"taskrunner/pkg/graph/task"
	"taskrunner/pkg/log"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
)

const (
	_taskLogCacheLen = 2048 * 1024
	_taskLogLevel    = logrus.DebugLevel
)

// NewPlan make plan for the job
func (e *Executor) NewPlan(ctx context.Context, jb *trait.JobRecord) (*graph.Plan, *trait.Error) {
	sid := jb.Target.SID
	s, err := e.getSystemContext(ctx, sid)
	if err != nil {
		e.Store.Log.Errorf("get the system %d context error: %s", sid, err.Error())
		return nil, err
	}

	target := jb.Target
	app, err := e.GetAPP(ctx, target.AID)
	if err != nil {
		e.Log.Errorf("get the application error: %s", err.Error())
		return nil, err
	}
	target.Application = *app

	p, external, err := e.newJobs(jb, *s)
	if err != nil {
		e.Log.Errorf("create job plan error: %s", err.Error())
		return nil, err
	}
	e.Log.Debugf("external node num: %d", len(external))
	err = e.gotExternalAttribute(ctx, sid, external)
	return p, err
}

// CheckDepComponents check dependence component instance
func (e *Executor) CheckDepComponents(ctx context.Context, app *trait.Application, sid int) *trait.Error {
	ts := make([]trait.Task, 0, len(app.Component))
	for _, c := range app.Component {
		ts = append(ts, &task.Base{
			ComponentInsData: &trait.ComponentInstance{
				ComponentInstanceMeta: trait.ComponentInstanceMeta{
					Component: c.ComponentNode,
				},
			},
		})
	}
	_, external, err := graph.NewFromGraph(app.Graph, ts)
	if err != nil {
		e.Log.Errorf("create task plan from application %s for check dependence, error: %s", app.AName, err.Error())
		return err
	}
	return e.gotExternalAttribute(ctx, sid, external)
}

func (e *Executor) newJobs(jb *trait.JobRecord, system cluster.SystemContext) (*graph.Plan, []*task.Base, *trait.Error) {
	ts, err := task.NewTasks(&jb.Target.Application, &system)
	if err != nil {
		e.Log.Errorf("create task from application error: %s", err.Error())
		return nil, nil, err
	}

	for _, t := range ts {
		c := t.Component()
		cins := jb.Target.ComponentInsExistedOrCreate(*c)
		cins.AppConfig = jb.Target.AppConfig

		if err := t.SetComponentIns(cins); err != nil {
			e.Log.Errorf("set component instance data for component [%s:%s] error, %s", c.Name, c.Version, err.Error())
			return nil, nil, err
		}
	}
	e.Log.Debugf("make plan for job %d, has %d components", jb.ID, len(ts))
	plan, external, err := graph.NewFromGraph(jb.Target.Graph, ts)
	return plan, external, err
}

func (e *Executor) gotExternalAttribute(ctx context.Context, sid int, ts []*task.Base) *trait.Error {
	for _, t := range ts {
		c, err := e.Store.GetWorkComponentIns(ctx, sid, *t.Component())
		if err != nil {
			if trait.IsInternalError(err, trait.ErrNotFound) {
				err = &trait.Error{
					Internal: trait.ErrComponentNotFound,
					Detail:   t.ComponentInsData.Component.Name,
					Err:      fmt.Errorf("get component [%s] in system [%d] error:%s", t.ComponentInsData.Component.Name, sid, err.Error()),
				}
			}
			e.Store.Log.Error(err.Error())
			return err
		}
		// warn: don't replace the t.ComponentInsData with c, pointer c may use by other one
		cins := t.ComponentInsData
		cins.AIID = c.AIID
		cins.Acid = c.Acid
		cins.APPName = c.APPName
		cins.Status = c.Status
		cins.Config = c.Config
		cins.Config = c.Config
		cins.Attribute = c.Attribute
		cins.Timeout = c.Timeout
		cins.Revission = c.Revission
		cins.CID = c.CID

		if semver.Compare(semver.MajorMinor("v"+c.Component.Version), semver.MajorMinor("v"+t.ComponentInsData.Component.Version)) == -1 {
			err := fmt.Errorf("the cur component [%s:%s] less than expect [%s]",
				c.Component.Name, c.Component.Version, t.ComponentInsData.Component.Version)
			e.Store.Log.Error(err.Error())
			return &trait.Error{
				Err:      err,
				Internal: trait.ErrComponentVersionLess,
				Detail:   c.Component,
			}
		}
	}
	return nil
}

func (e *Executor) getDepNode(ctx context.Context, job *trait.JobRecord) ([]int, *trait.Error) {
	// index use to unique component instance id
	index := map[int]bool{}
	for _, cins := range job.Target.Components {
		if cins.Component.ComponentDefineType == component.ComponentBaseType {
			continue
		}
		index[cins.CID] = false
		nodes, err := e.Store.GetPointTo(ctx, cins.CID)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			err = nil
			nodes = nil
		}
		if err != nil {
			e.Store.Log.Errorf("get component instance point to %s error: %s", cins.Component.Name, err.Error())
			return nil, err
		}
		for _, cid := range nodes {
			if _, ok := index[cid]; !ok {
				index[cid] = true
			}
		}
	}
	ids := make([]int, 0, len(index))
	for k, is := range index {
		if is {
			ids = append(ids, k)
		}
	}
	return ids, nil
}

// heavyReupdateComponent update the component with current like realtime dependent  without cache
func (e *Executor) heavyReupdateComponent(ctx, jobContext context.Context, cid int, job *trait.JobRecord) (err *trait.Error) {
	l := log.NewTaskLogger(e.Log, _taskLogLevel, _taskLogCacheLen)

	logErr := func(cins *trait.ComponentInstance, err *trait.Error) {
		jobLogRecord := trait.JobLog{
			JID:       job.ID,
			CID:       cid,
			Code:      err.Internal,
			Msg:       string(l.Bytes()),
			Timestamp: int(time.Now().Unix()),
		}
		l.Reset()
		if cins != nil {
			jobLogRecord.CID = cins.CID
			jobLogRecord.AIID = cins.AIID
			jobLogRecord.Aname = cins.APPName
			jobLogRecord.Cname = cins.Component.Name
		}

		_ = utils.RetryN(ctx, func() (bool, *trait.Error) {
			if err0 := e.Store.InsertJobLog(ctx, jobLogRecord); err0 != nil {
				e.Store.Log.Warnf("log job fail into store error: %s", err0.Error())
				return true, err0
			}
			return false, nil
		}, 3, 500*time.Millisecond)
	}

	var cins *trait.ComponentInstance
	if err := utils.RetryN(ctx, func() (bool, *trait.Error) {
		var err *trait.Error
		cins, err = e.Store.GetComponentIns(jobContext, cid)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			// the instance has been delete, no need to upgrade
			e.Store.Log.Warnf("the component instance %d not exists, ignore upgrade deps in this component", cid)
			return false, nil
		}
		if err != nil {
			e.Store.Log.Errorf("get component instance for upgrade dep fail: %s, retry later", err.Error())
			logErr(nil, err)
			return true, nil
		}
		return false, nil
	}, 3, 500*time.Millisecond); err != nil {
		e.Store.Log.Errorf("get component instance for upgrade dep fail: %s", err.Error())
		logErr(nil, err)
	}

	if cins == nil {
		return nil
	}

	e.Store.Log.Debugf("update external parent component: %s", cins.Component.Name)
	cins.StartTime = int(time.Now().Unix())

	getWorkCompnent := func() (*trait.ComponentInstance, bool, *trait.Error) {
		// marco for get compoent return true when break
		var cins0 *trait.ComponentInstance
		exit := false
		if err0 := utils.RetryN(ctx, func() (bool, *trait.Error) {
			cins0, err = e.Store.GetWorkComponentIns(jobContext, cins.System.SID, cins.Component)
			if trait.IsInternalError(err, trait.ErrNotFound) {
				// the work instance has been removed, no need to upgrade
				e.Store.Log.Warnf("the work component %s in system %d  not exists, ignore upgrade deps in this component", cins.Component.Name, cins.System.SID)
				exit = true
				return false, nil
			}
			if err != nil {
				e.Store.Log.Errorf("get system %d component %s installer for upgrade deps error: %s", cins.System.SID, cins.Component.Name, err.Error())
				return true, err
			}
			if cins.Status != trait.AppSucessStatus {
				exit = true
				e.Store.Log.Warnf("the component instance %d has install fail by other job, ignore it", cins.CID)
				return false, nil
			}
			exit = false
			return false, nil
		}, 3, 500*time.Millisecond); err0 != nil {
			l.Errorf("get system %d component %s installer for upgrade deps error: %s", cins.System.SID, cins.Component.Name, err.Error())
			logErr(nil, err0)
			return cins0, exit, err0
		}

		return cins0, exit, nil
	}
	unlockComponent := func() *trait.Error {
		ctx0, cancel := trait.WithTimeoutCauseContext(context.TODO(), 30*time.Second, nil)
		defer cancel()
		err0 := utils.RetryN(ctx0, func() (bool, *trait.Error) {
			ctx1, cancel := trait.WithTimeoutCauseContext(context.TODO(), 5*time.Second, &trait.Error{
				Internal: trait.ECTimeout,
				Err:      fmt.Errorf("warn!!! unlock component lock aiid: %d, sid:%d, cname: %s timeout", cins.AIID, cins.System.SID, cins.Component.Name),
				Detail:   "heavyReupdateComponent",
			})
			defer cancel()
			if err0 := e.Store.UnlockComponent(ctx1, cins.System.SID, job.Target.ID, cins.Component); err0 != nil {
				return true, err0
			}
			return false, nil
		}, 10, 500*time.Millisecond)

		if err0 != nil {
			l.Errorf("warn!!! unlock component lock aiid: %d, sid:%d, cname: %s, error: %s", cins.AIID, cins.System.SID, cins.Component.Name, err0.Error())
		}
		return err0
	}

	for count := 0; ; count++ {
		// first time, workcomponent may is cahche, need reload.
		// second time, workcomponent must has been update by other job. no need reupdate
		if err = e.Store.LockComponent(jobContext, cins.System.SID, job.Target.ID, cins.Component); err != nil {
			e.Store.Log.Errorf("job's application instance id: %d, system id: %d,  get component: %s, lock fail: %s",
				job.Target.ID, cins.System.SID, cins.Component.Name, err.Error())
			return
		}
		cins0, exit, err := getWorkCompnent()
		if err != nil || exit {
			defer func() {
				_ = unlockComponent()
			}()
			return err
		}
		if cins0.CID == cins.CID {
			break
		}

		e.Store.Log.Debugf("workcomponent %s has been refresh, try reload only once", cins.Component.Name)
		if err = unlockComponent(); err != nil {
			defer logErr(cins, err)
			return err
		}

		if count == 1 {
			e.Store.Log.Debugf("the component %s has been reupdate by other job, ignore it", cins.Component.Name)
			return nil
		}

		cins = cins0
	}
	defer func() {
		_ = unlockComponent()
	}()

	{
		var ains *trait.ApplicationInstance
		if err0 := utils.RetryN(ctx, func() (bool, *trait.Error) {
			ains, err = e.Store.GetAPPIns(ctx, cins.AIID)
			if trait.IsInternalError(err, trait.ErrNotFound) {
				// the component install task info has been removed, we think it has been uninstall
				e.Store.Log.Warnf("the application instance %d not exists, ignore upgrade deps in this component", cins.Acid)
				return false, nil
			}
			if err != nil {
				e.Store.Log.Errorf("get application %d component %s installer for upgrade deps error: %s", cins.AIID, cins.Component.Name, err.Error())
				return true, err
			}
			return false, nil
		}, 3, 500*time.Millisecond); err0 != nil {
			l.Errorf("get application instance %d component %s installer for upgrade deps error: %s", cins.AIID, cins.Component.Name, err.Error())
			logErr(cins, err0)
			return err0
		}

		var cm *trait.ComponentMeta
		if err0 := utils.RetryN(ctx, func() (bool, *trait.Error) {
			cm, err = e.Store.GetAPPComponent(ctx, cins.Acid)
			if trait.IsInternalError(err, trait.ErrNotFound) {
				// the component install task info has been removed, we think it has been uninstall
				e.Store.Log.Warnf("the application component %d not exists, ignore upgrade deps in this component", cins.Acid)
				return false, nil
			}
			if err != nil {
				e.Store.Log.Errorf("get application %d component %s installer for upgrade deps error: %s", cins.AIID, cins.Component.Name, err.Error())
				return true, err
			}
			return false, nil
		}, 3, 500*time.Millisecond); err0 != nil {
			l.Errorf("get application %d component %s installer for upgrade deps error: %s", cins.AIID, cins.Component.Name, err.Error())
			logErr(cins, err0)
			return err0
		}
		if cm == nil {
			return
		}

		var sysContext *cluster.SystemContext
		if err0 := utils.RetryN(ctx, func() (bool, *trait.Error) {
			sysContext, err = e.getSystemContext(ctx, cins.System.SID)
			if trait.IsInternalError(err, trait.ErrNotFound) {
				// the component install task info has been removed, we think it has been uninstall
				e.Store.Log.Warnf("the system %d not exists, ignore upgrade deps in this system", cins.System.SID)
				err = nil
				return false, nil
			}
			if err != nil {
				e.Store.Log.Errorf("get systemcontext %d for upgrade deps error: %s", cins.System.SID, err.Error())
				return true, err
			}
			return false, nil
		}, 3, 500*time.Millisecond); err0 != nil {
			l.Errorf("get systemcontext %d for upgrade deps error: %s", cins.System.SID, err0.Error())
			logErr(cins, err0)
			return err0
		}
		if sysContext == nil {
			return
		}
		var children []*trait.ComponentInstance

		if err = utils.RetryN(ctx, func() (bool, *trait.Error) {
			ids, err0 := e.Store.GetPointFrom(ctx, cins.CID)
			if err0 != nil {
				e.Store.Log.Errorf("get children node instance error: %s", err.Error())
				return true, err0
			}
			cs := make([]*trait.ComponentInstance, 0, len(ids))
			for _, cid := range ids {
				child, err0 := e.Store.GetComponentIns(ctx, cid)
				if trait.IsInternalError(err, trait.ErrNotFound) {
					e.Log.Warnf("the component instance is not exists ")
					return false, err0
				}
				if err0 != nil {
					e.Store.Log.Errorf("get children node instance error: %s", err0.Error())
					return true, err0
				}
				cs = append(cs, child)
			}
			children = cs
			return false, nil
		}, 3, 500*time.Millisecond); err != nil {
			l.Errorf("get children node for heavy upgrade component error: %s", err.Err)
			logErr(cins, err)
			return err
		}

		ts, err0 := task.NewTask(cm, sysContext)
		err = err0
		if err != nil {
			l.Error(err)
			logErr(cins, err)
			return
		}
		cins.AppConfig = ains.AppConfig
		ts.SetTopology(children)
		if err := ts.SetComponentIns(cins); err != nil {
			e.Store.Log.Errorf("set task for component instance error: %s, ignore it", err.Error())
			return nil
		}

		ts.WithLog(l)
		err = ts.Install(jobContext)
		if err != nil {
			l.Errorf("upgrade component %s who depend application %d, error: %s", cins.Component.Name, job.Target.AID, err.Error())
			logErr(cins, err)
			cins.Status = trait.AppFailStatus
		} else {
			cins.Status = trait.AppSucessStatus
		}
		cins.EndTime = int(time.Now().Unix())
		if err0 = e.Store.UpdateComponentInsStatus(ctx, cins.CID, cins.Status, cins.Revission, cins.StartTime, cins.EndTime); err0 != nil {
			l.Errorf("please check db store, store component sucess status error:%s", err0.Error())
			logErr(cins, err0)
			return
		}
		cins.Revission++
	}

	return err
}

func (e *Executor) heavyUpdateParent(ctx, jobContext context.Context, job *trait.JobRecord) *trait.Error {
	ids, err := e.getDepNode(jobContext, job)
	if err != nil {
		return err
	}
	e.Log.Debugf("job %d heavyUpdateParent need update %d external components", job.Target.ID, len(ids))
	wg := &sync.WaitGroup{}
	parallel := e.parallel
	wg.Add(parallel)

	ctx0, cancel := context.WithCancel(jobContext)
	defer cancel()

	ch := make(chan int)
	cherr := make(chan *trait.Error, parallel)
	count := int64(0)
	errCount := int64(0)
	sendErr := func(err *trait.Error) {
		atomic.AddInt64(&errCount, 1)
		cherr <- err
		cancel()
	}

	go func() {
		defer close(ch)
		for _, cid := range ids {
			select {
			case <-ctx0.Done():
				return
			case ch <- cid:
			}
		}
	}()

	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx0.Done():
					return
				case cid, ok := <-ch:
					if !ok {
						return
					}
					atomic.AddInt64(&count, 1)
					if err := e.heavyReupdateComponent(ctx, jobContext, cid, job); err != nil {
						sendErr(err)
						return
					}
				}
			}
		}()
	}

	wg.Wait()
	close(cherr)
	e.Log.Debugf("heavyReupdateComponent job %d, exeute %d, error %d", job.ID, count, errCount)
	// only get one error
	errCountInt := int(errCount)
	for i := 0; i < errCountInt; i++ {
		oneErr, ok := <-cherr
		if !ok {
			break
		}
		if oneErr != nil {
			err = oneErr
			if trait.IsInternalError(err, trait.ECExit) || trait.IsInternalError(err, trait.ECJobCancel) {
				return err
			}
		}
	}
	if ctxErr := ctx.Err(); ctxErr != nil && err == nil {
		return ctx.Err().(*trait.Error)
	}
	return err
}

func (e *Executor) upgradeComponentStage(ctx context.Context, jc *jobControl) (jobStatus int, err *trait.Error) {
	job := jc.job

	jobStatus = trait.AppFailStatus

	plan, err := e.NewPlan(ctx, job)
	if trait.IsInternalError(err, trait.ErrNotFound) {
		e.Store.Log.Errorf("make application plan error for missing depence: %s", err.Error())
		jobStatus = trait.AppFailMissStatus
	}
	if err != nil {
		jc.Lock()
		job.Target.Status = jobStatus
		defer jc.Unlock()

		e.Store.Log.Error("make plan error, stop the job")
		if err0 := e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, 0, job.Target.StartTime, int(time.Now().Unix())); err0 != nil {
			e.Store.Log.Errorf("stop job fail when make plan error: %s", err0.Error())
			return
		}
		return
	}

	jc.Lock()
	if job.Target.Status == trait.AppStopedStatus {
		// job has been stoped
		jc.Unlock()
		return
	}
	job.Target.Status = trait.AppDoingStatus
	// update store first befor update memory, so that don't overwrite stopped status
	// TODO retry when error
	if err = e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, e.id, job.Target.StartTime, -1); err != nil {
		jc.Unlock()
		e.Store.Log.Errorf("start waiting job fail when change job status: %s", err.Error())
		return
	}
	jc.Unlock()

	wg := sync.WaitGroup{}
	// consider to disk io and net io, parallel 1-> 3
	parallel := e.parallel
	wg.Add(parallel)

	e.Store.Log.Debugf("execute job %d plan parrell %d", job.ID, parallel)
	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			_ = e.executeTask(ctx, jc, plan)
		}()
	}

	finished := make(chan int)
	defer close(finished)
	go func() {
		// control routine will stop executeTask routine and stop by parent routine
		select {
		case <-ctx.Done():
			e.Store.Log.Infof("receive main routine's exit signal, try to stop plan execute")
			plan.Close()
		case <-finished:
			// exit control routine
		case <-jc.ctx.Done():
			// cancel job
			if trait.IsInternalError(jc.ctx.Err(), trait.ECJobCancel) {
				e.Store.Log.Infof("recevice job %d cancel signal, try to stop plan execute", job.ID)
			} else {
				e.Store.Log.Infof("recevice engine exit signal %s, try to stop plan execute", jc.ctx.Err().Error())
			}
			plan.Close()
		}
	}()

	wg.Wait()
	e.Store.Log.Infof("job %d plan execute routine end", job.ID)

	if trait.IsInternalError(ctx.Err(), trait.ECExit) {
		e.Store.Log.Info("main routine context end, exit quick")
		return jobStatus, ctx.Err().(*trait.Error)
	}

	jc.Lock()
	defer jc.Unlock()
	// success
	jobStatus = trait.AppUpdatedComponentStatus
	count := 0
	for _, i := range job.Target.Components {
		count++
		if i.Status != trait.AppSucessStatus && i.Status != trait.AppIgnoreStatus {
			jobStatus = trait.AppFailStatus
			job.Target.EndTime = int(time.Now().Unix())
			break
		}
	}

	onwer := 0
	if trait.IsInternalError(jc.ctx.Err(), trait.ECJobCancel) || trait.IsInternalError(jc.ctx.Err(), trait.ECExit) {
		jobStatus = trait.AppStopedStatus
	} else if jobStatus == trait.AppUpdatedComponentStatus {
		// all component success, next stage
		jobStatus = trait.AppUpdatedComponentStatus
		onwer = e.id
	} else {
		jobStatus = trait.AppFailStatus
		job.Target.EndTime = int(time.Now().Unix())
	}

	if jobStatus == trait.AppStopedStatus {
		if err := utils.RetryN(ctx, func() (bool, *trait.Error) {
			if err := e.Store.UnlockJobComponent(ctx, job.ID); err != nil {
				e.Log.Errorf("UnlockJobComponent %d error: %s", job.ID, err.Error())
				return true, err
			}
			return false, nil
		}, 100, time.Second); err != nil {
			e.Store.Log.Error(err.Error())
			return jobStatus, err
		}
	}

	if err = e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, jobStatus, onwer, job.Target.StartTime, job.Target.EndTime); err != nil {
		e.Store.Log.Errorf("update application instance status for upgrade stage error: %s", err.Error())
		return
	}
	err0 := trait.UnwrapError(jc.ctx.Err())
	if err0 != nil {
		err = err0
	} else if jobStatus != trait.AppUpdatedComponentStatus {
		err = &trait.Error{
			Internal: trait.ECNULL,
			Err:      context.Canceled,
			Detail:   fmt.Sprintf("the job %d fail with status %d and stop , see pre log for detail", job.Target.Status, job.Target.ID),
		}
	}

	e.Store.Log.Debugf("execute update job %d component upgrade stage end, update job into status %d", job.Target.ID, jobStatus)

	return
}

func (e *Executor) CleanJob(ctx context.Context, s trait.ApplicationInsWriter, job *trait.JobRecord) *trait.Error {
	if err := s.UnlockJobComponent(ctx, job.ID); err != nil {
		e.Log.Errorf("UnlockJobComponent %d error: %s", job.ID, err.Error())
		return err
	}

	if err := s.UnlockApp(ctx, job.Target.SID, job.ID, job.Target.AName); err != nil {
		e.Store.Log.Tracef("retry: unlock application: %s, jid: %d sid: %d, error: %s",
			job.Target.AName, job.ID, job.Target.SID, err.Error())
		return err
	}

	return nil
}

func isFinish(status int) bool {
	for _, s := range trait.JobDoingStauts {
		if s == status {
			return false
		}
	}
	return true
}

// finishJob 对进入结束状态的任务必须调用的函数，该函数将会清理任务相关锁资源与设置任务状态。
// 成功状态将会使得任务被设置进入工作表。
func (e *Executor) finishJob(ctx context.Context, jc *jobControl) *trait.Error {
	job := jc.job
	status := jc.GetStatus()
	if isFinish(status) {
		// 终止状态强制清理和设置为终止态,过程状态仍使用上下文
		ctx = context.Background()
	}
	for {
		// 自多次重试改为循环，防止数据库长时间崩溃后需要引入手工修复
		// change job status
		changeStatus := func() *trait.Error {
			tx, err := e.Store.Begin(ctx)
			if err != nil {
				e.Store.Log.Errorf("start transaction error when change job status: %s", err.Error())
				return err
			}

			// 最终清理
			if err := e.CleanJob(context.Background(), tx, job); err != nil {
				e.rollbackWithLog(tx)
				return err
			}

			if status == trait.AppSucessStatus {

				if job.Current != nil {
					e.Store.Log.Debugf("layoff application instance %d from work", job.Current.ID)
					if err := tx.LayOffAPPIns(ctx, job.Target); err != nil {
						e.Store.Log.Errorf("layoff old application instance %d error, err: %s", job.Current.ID, err)
						e.rollbackWithLog(tx)()
						return err
					}
				}

				e.Store.Log.Debugf("set applicatin instance %d into work", job.Target.ID)
				if err := tx.WorkAppIns(ctx, job.Target); err != nil {
					e.Store.Log.Errorf("work application instance %d error, err: %s", job.Target.ID, err)
					e.rollbackWithLog(tx)()
					return err
				}
			}

			if err := tx.UpdateAPPInsStatus(ctx, job.Target.ID, status, 0, job.Target.StartTime, job.Target.EndTime); err != nil {
				e.Store.Log.Errorf("start waiting job fail when change job status: %s", err.Error())
				e.rollbackWithLog(tx)()
				return err
			}

			if err := tx.Commit(); err != nil {
				e.Store.Log.Errorf("commit job status error: %s", err.Error())
				e.rollbackWithLog(tx)
				return err
			}

			return nil
		}

		if err := changeStatus(); err != nil {
			if trait.IsInternalError(err, trait.ECExit) {
				// 进程退出，无需处理，进程恢复时进行的灾难恢复会覆盖该任务执行
				return err
			}
			// 3s后重试
			e.Log.Warnf("change job status error: %s, will retry after 3s", err.Error())
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	return nil
}

func (e *Executor) executeJob(ctx context.Context, jc *jobControl) *trait.Error {
	job := jc.job
	cleanOk := false
	defer func() {
		// 不进行重复清理
		if cleanOk {
			return
		}
		for {
			// 任务结束除进程退出外，必须释放锁与任务相关资源，避免后续进程并发问题。
			// 该defer清理为最终保障，任务失败未释放锁也会导致并非锁无法释放与后续获取问题，
			// 且进入该循环的场景大都为任务失败场景，因此将强制要求清理，忽略进程退出信号
			if err := e.CleanJob(context.Background(), e.Store, job); err != nil {
				if trait.IsInternalError(err, trait.ECExit) {
					return
				}
				e.Store.Log.Errorf("unlock application: %s, jid: %d sid: %d, error: %s",
					job.Target.AName, job.ID, job.Target.SID, err.Error())
				time.Sleep(2 * time.Second)
			} else {
				return
			}
		}
	}()

	// app lock avoid loop locker acquire
	if err := utils.RetryN(ctx, func() (bool, *trait.Error) {
		err := e.LockApp(ctx, job.Target.SID, job.ID, job.Target.AName)
		if trait.IsInternalError(err, trait.ECExit) ||
			trait.IsInternalError(err, trait.ECJobCancel) {
			return false, err
		} else if err != nil {
			return true, err
		}
		return false, nil
	}, 10, 500*time.Millisecond); err != nil {
		e.Store.Log.Errorf("get application %s locker for job %d in system %d error: %s",
			job.Target.AName, job.ID, job.Target.SID, err.Error())
		return err
	}

	// release job hole component lock avoid loop locker acquire
	if err := utils.RetryN(ctx, func() (bool, *trait.Error) {
		err := e.UnlockJobComponent(ctx, job.ID)
		if trait.IsInternalError(err, trait.ECTimeout) {
			return false, err
		} else if err != nil {
			return true, err
		}
		return false, nil
	}, 10, 500*time.Millisecond); err != nil {
		return err
	}

	jc.Lock()
	if job.Target.Status == trait.AppStopedStatus {
		// stop
		jc.Unlock()
		e.Store.Log.Infof("receive stoped job %d, ignore", jc.job.ID)
		return nil
	}

	if job.Target.Status == trait.AppStopingStatus {
		defer jc.Unlock()
		e.Store.Log.Infof("receive a job %d need stop before start, stop it", jc.job.ID)
		if err := e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, trait.AppStopedStatus, 0, job.Target.StartTime, int(time.Now().Unix())); err != nil {
			e.Store.Log.Errorf("change job %d status into stopped error: %s ", job.Target.ID, err.Error())
			return err
		}
		return nil
	}

	if jc.job.Target.OType == trait.JobDeleteOType {
		jc.Unlock()
		// 删除任务
		err := e.deleteJob(ctx, jc)
		if err != nil {
			// warn 存在状态为运作状态且退出风险,因此调用finish设置状态
			jc.SetStatus(trait.AppFailUninstallStatus)
			_ = e.finishJob(ctx, jc)
			cleanOk = true
		}
		return err
	}
	jc.Unlock()

	// 后续代码不做变更,仍未安装/更新流程
	// 卸载流程-first(不保持旧逻辑，先卸载)
	if !jc.job.Target.Trait.RetainOrder {
		if err := e.deleteTask(ctx, jc, nil); err != nil {
			// warn 存在状态为运作状态且退出风险,因此调用finish设置状态
			jc.SetStatus(trait.AppFailUninstallStatus)
			_ = e.finishJob(ctx, jc)
			cleanOk = true
			return err
		}
	}

	// 安装流程
	jobStatus, err := e.upgradeComponentStage(ctx, jc)
	if err != nil {
		jc.SetStatus(trait.AppFailStatus)
		_ = e.finishJob(ctx, jc)
		cleanOk = true
		return err
	}
	jc.Lock()
	// stop
	if jobStatus == trait.AppStopedStatus || jobStatus == trait.AppFailStatus {
		jc.Unlock()
		return nil
	}

	jc.Unlock()

	// 卸载流程-finally（保持旧逻辑，后卸载）
	if jc.job.Target.Trait.RetainOrder {
		if err := e.deleteTask(ctx, jc, nil); err != nil {
			// warn 存在状态为运作状态且退出风险,因此调用finish设置状态
			jc.SetStatus(trait.AppFailUninstallStatus)
			_ = e.finishJob(ctx, jc)
			cleanOk = true
			return err
		}
	}

	{
		jobStatus = trait.AppSucessStatus
		jc.SetStatus(jobStatus)
	}

	if job.Target.Trait.UpgradeParent {
		if err := e.heavyUpdateParent(ctx, jc.ctx, job); err != nil {
			jobStatus = trait.AppUpgradeParentComponentFailStatus
			if trait.IsInternalError(err, trait.ECJobCancel) {
				jobStatus = trait.AppStopedStatus
			}
			jc.SetStatus(jobStatus)
			e.Store.Log.Errorf("heavy upgrade parent node error %s", err.Error())
		}
	}

	job.Target.EndTime = int(time.Now().Unix())

	_ = e.finishJob(ctx, jc)
	cleanOk = true
	if job.Target.Status != trait.AppSucessStatus {
		err = &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("job %d execute fail, with status %d", job.ID, job.Target.Status),
			Detail:   job.Target.Status,
		}
		return err
	}

	return nil
}

func (e *Executor) deleteJob(ctx context.Context, jc *jobControl) *trait.Error {
	jc.Lock()
	// 伪造删除任务
	comIns := jc.job.Target.Components
	jc.job.Target.Components = nil
	if jc.job.Current == nil || jc.job.Current.ID == -1 {
		jc.job.Current = jc.job.Target
	}
	jc.Unlock()
	// app, err := e.GetAPP(ctx, jc.job.Current.AID)
	// if err != nil {
	// 	e.Log.Errorf("get the application error: %s", err.Error())
	// 	return err
	// }
	job := jc.job

	e.Store.Log.Infof("delete job has %d component", len(comIns))
	if err := e.deleteTask(ctx, jc, comIns); err != nil {
		e.Store.Log.Errorf("delete job error: %s", err.Error())
		if err0 := e.UpdateAPPInsStatus(
			ctx, job.Target.ID, trait.AppFailStatus,
			e.id, job.Target.StartTime, int(time.Now().Unix())); err0 != nil {
			e.Store.Log.Errorf("change delete job status into fail error: %s", err.Error())
			return err0
		}
		return err
	}

	job.Target.Status = trait.AppSucessStatus
	// change job status
	tx, err := e.Store.Begin(ctx)
	if err != nil {
		e.Store.Log.Errorf("start transaction error when change job status: %s", err.Error())
		return err
	}

	if job.Current != nil {
		e.Store.Log.Debugf("layoff application instance %d from work", job.Current.ID)
		if err := tx.LayOffAPPIns(ctx, job.Target); err != nil {
			e.Store.Log.Errorf("layoff old application instance %d error, err: %s", job.Current.ID, err)
			e.rollbackWithLog(tx)()
			return err
		}
	}

	e.Store.Log.Debugf("layoff applicatin instance %d from work", job.Target.ID)
	if err := tx.LayOffAPPIns(ctx, job.Target); err != nil {
		e.Store.Log.Errorf("layoff old application instance %d error, err: %s", job.Current.ID, err)
		e.rollbackWithLog(tx)()
		return err
	}

	if err := tx.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, 0, job.Target.StartTime, int(time.Now().Unix())); err != nil {
		e.Store.Log.Errorf("start waiting job fail when change job status: %s", err.Error())
		e.rollbackWithLog(tx)()
		return err
	}

	if err := tx.Commit(); err != nil {
		e.Store.Log.Errorf("commit job status error: %s", err.Error())
		return err
	}
	return nil
}

// logErr log err into database, caller nomally ignore error
// TODO move log into go routine worker
func (e *Executor) logErr(ctx context.Context, jid int, cins *trait.ComponentInstance, err *trait.Error, log *log.TaskLogger) *trait.Error {
	jobLogRecord := trait.JobLog{
		JID:       jid,
		Code:      err.Internal,
		Msg:       string(log.Bytes()),
		Timestamp: int(time.Now().Unix()),
	}
	log.Reset()
	if cins != nil {
		jobLogRecord.CID = cins.CID
		jobLogRecord.AIID = cins.AIID
		jobLogRecord.Aname = cins.APPName
		jobLogRecord.Cname = cins.Component.Name
	}

	return utils.RetryN(ctx, func() (bool, *trait.Error) {
		if err0 := e.Store.InsertJobLog(ctx, jobLogRecord); err0 != nil {
			e.Store.Log.Warnf("log job fail into store error: %s", err0.Error())
			return true, err0
		}
		return false, nil
	}, 3, 500*time.Millisecond)
}

func (e *Executor) executeTask(ctx context.Context, jc *jobControl, p *graph.Plan) *trait.Error {
	// job := jc.job
	jobContext := jc.ctx
	log := e.newJobLogger()
	for {
		log.Reset()
		taskNode := p.NextBlock()
		if taskNode == nil {
			return nil
		}
		t := taskNode.Task
		cins := t.ComponentIns()

		e.Store.Log.Debugf("execute task node %#v", t.Component())
		if _, ok := t.(*task.Base); ok {
			// ignore external node
			e.Store.Log.Debugf("ignore external node %s", t.Component().Name)
			p.Done(taskNode)
			continue
		}

		setwork := func(status int) *trait.Error {
			if err := utils.RetryN(ctx, func() (bool, *trait.Error) {
				if err := e.setWorkComponentIns(ctx, taskNode, status); err != nil {
					return true, err
				}
				return false, nil
			}, 10, 500*time.Millisecond); err != nil {
				cins.Status = trait.AppFailStatus
				log.Errorf("set component instance into work status error: %s", err.Err)
				_ = e.logErr(ctx, jc.job.ID, cins, err, log)
				defer p.Close()
				return err
			}
			cins.Status = status
			p.Done(taskNode)
			return nil
		}

		// 1. get task child config
		// 2. install task
		// 3. set task into work component table
		if err := func() *trait.Error {
			status := cins.Status

			e.Store.Log.Debugf("get component lock. aiid: %d, sid: %d, cname: %s", cins.AIID, cins.System.SID, cins.Component.Name)
			if err := e.Store.LockComponent(ctx, cins.System.SID, cins.AIID, cins.Component); err != nil {
				log.Errorf("get component lock aiid: %d, sid:%d, cname: %s, error: %s", cins.AIID, cins.System.SID, cins.Component.Name, err.Error())
				_ = e.logErr(ctx, jc.job.ID, cins, err, log)
				return err
			}
			e.Store.Log.Debugf("get component lock sucess. aiid: %d, sid: %d, cname: %s", cins.AIID, cins.System.SID, cins.Component.Name)

			// defer unlock component
			defer func() {
				ctx0, cancel := trait.WithTimeoutCauseContext(context.TODO(), 30*time.Second, nil)
				defer cancel()
				_ = utils.RetryN(ctx0, func() (bool, *trait.Error) {
					ctx1, cancel := trait.WithTimeoutCauseContext(context.TODO(), 5*time.Second, &trait.Error{
						Internal: trait.ECTimeout,
						Err:      fmt.Errorf("warn!!! unlock component lock aiid: %d, sid:%d, cname: %s timeout", cins.AIID, cins.System.SID, cins.Component.Name),
						Detail:   "executeTask",
					})
					defer cancel()
					if err0 := e.Store.UnlockComponent(ctx1, cins.System.SID, cins.AIID, cins.Component); err0 != nil {
						e.Store.Log.Errorf("warn!!! unlock component lock aiid: %d, sid:%d, cname: %s, error: %s", cins.AIID, cins.System.SID, cins.Component.Name, err0.Error())
						return true, err0
					}
					return false, nil
				}, 6, 500*time.Millisecond)
			}()

			if cins.Status == trait.AppIgnoreStatus || cins.Status == trait.AppSucessStatus {
				// the component task has been installed or can ignore
				if cins.StartTime == 0 {
					// ingore time not set
					cins.StartTime = int(time.Now().Unix())
				}
				if cins.EndTime == 0 {
					cins.EndTime = int(time.Now().Unix())
				}
				e.Store.Log.Infof("component %s status is %d, ignore the task execute, jump to work component stage", cins.Component.Name, cins.Status)
				return setwork(status)
			}

			// only record the time for execute
			cins.StartTime = int(time.Now().Unix())

			e.Store.Log.Debugf("execute plan task node component aid: %d, cid: %d, %s/%s:%s, status:%d",
				cins.AIID, cins.CID, cins.APPName, cins.Component.Name, cins.Component.Version, cins.Status)

			setCompoentfail := func(status int, err *trait.Error) {
				defer p.Close()
				cins.Status = status
				if err := utils.RetryN(ctx, func() (bool, *trait.Error) {
					if err0 := e.Store.UpdateComponentInsStatus(ctx, cins.CID, cins.Status, cins.Revission, cins.StartTime, int(time.Now().Unix())); err0 != nil {
						log.Errorf("please check db store, store component status error:%s", err0.Error())
						return true, err0
					}
					cins.Revission++
					return false, nil
				}, 10, 500*time.Millisecond); err != nil {
					log.Errorf("update component instance into database after retry still error : %s", err.Error())
				}
				_ = e.logErr(ctx, jc.job.ID, cins, err, log)
			}

			c := t.Component()
			topology := make([]*trait.ComponentInstance, 0, len(taskNode.Children))
			cur, err0 := e.Store.GetComponentIns(ctx, cins.CID)
			if err0 != nil {
				log.Errorf("get current component %s:%d instance error: %s", cins.Component.Name, cins.CID, err0.Error())
				setCompoentfail(trait.AppFailStatus, err0)
				return err0
			}
			for _, children := range taskNode.Children {
				// if real revission != cache instance revission,the component must has been upgrade by other job
				// so we need to upgrade external node in childrens
				if cur.Revission != cins.Revission && children.Component().ComponentDefineType == component.ComponentBaseType {
					cins := children.ComponentIns()
					c, err := e.Store.GetWorkComponentIns(ctx, cins.System.SID, cins.Component)
					if err != nil {
						defer p.Close()
						status := trait.AppFailStatus

						if trait.IsInternalError(err, trait.ErrNotFound) {
							err = &trait.Error{
								Internal: trait.ErrComponentNotFound,
								Detail:   cins.Component.Name,
								Err:      fmt.Errorf("get component [%s] in system [%d] error:%s", cins.Component.Name, cins.System.SID, err.Error()),
							}
							status = trait.AppFailMissStatus
						}

						log.Errorf("get current work component instance error: %s", err.Error())
						setCompoentfail(status, err)
						return err
					}
					if err = children.SetComponentIns(c); err != nil {
						defer p.Close()
						log.Errorf("set component instance for child error %s", err.Error())
						setCompoentfail(trait.AppFailStatus, err)
						return err
					}
				}
				childIns := children.ComponentIns()
				e.Store.Log.Debugf("node %s depends %s attribute : %v", cins.Component.Name, childIns.Component.Name, childIns.Attribute)
				topology = append(topology, childIns)
			}
			// jc.Lock()
			t.SetTopology(topology)
			// jc.Unlock()

			e.Log.Debugf("execute task timeout: %ds", cins.Timeout)
			t.WithLog(log)
			if err0 := t.Install(jobContext); err0 != nil {

				// jc.Lock()
				// try to close the plan, signal all other worker
				log.Errorf("install %s:%s error: %s", c.Name, c.Version, err0.Error())
				defer p.Close()
				setCompoentfail(trait.AppFailStatus, err0)
				return err0
			}
			status = trait.AppSucessStatus
			cins.EndTime = int(time.Now().Unix())
			return setwork(status)
		}(); err != nil {
			defer p.Close()
			return err
		}

	}
}

// deleteTask delete the old task
func (e *Executor) deleteTask(ctx context.Context, jc *jobControl, mustClean []*trait.ComponentInstance) *trait.Error {
	job := jc.job
	if job.Current == nil || job.Current.ID == -1 {
		return nil
	}
	// 必须清理项增加旧任务中组件,在后续代码中为目标任务中组件去重
	mustClean = append(mustClean, job.Current.Components...)

	e.Store.Log.Infof("job %d start delete the old component", job.ID)
	// update status
	jc.Lock()
	job.Target.Status = trait.AppDeleteingOldComponentStatus
	jc.Unlock()

	if err := e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, e.id, job.Target.StartTime, -1); err != nil {
		e.Store.Log.Errorf("change application status into deleting error: %s", err.Error())
		return err
	}
	l := log.NewTaskLogger(e.Store.Log, _taskLogLevel, _taskLogCacheLen)

	setJobUninstallFail := func() {
		// try update, not must
		jc.Lock()
		job.Target.Status = trait.AppFailUninstallStatus
		jc.Unlock()
		_ = utils.RetryN(ctx, func() (bool, *trait.Error) {
			if err := e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, 0, job.Target.StartTime, int(time.Now().Unix())); err != nil {
				e.Store.Log.Errorf("change application status into uninstall fail error: %s", err.Error())
				return true, err
			}
			return false, nil
		}, 3, 500*time.Millisecond)
	}

	logErr := func(cins *trait.ComponentInstance, err *trait.Error) {
		jobLogRecord := trait.JobLog{
			JID:       job.ID,
			Code:      err.Internal,
			Msg:       string(l.Bytes()),
			Timestamp: int(time.Now().Unix()),
		}
		l.Reset()
		if cins != nil {
			jobLogRecord.CID = cins.CID
			jobLogRecord.AIID = cins.AIID
			jobLogRecord.Aname = cins.APPName
			jobLogRecord.Cname = cins.Component.Name
		}

		_ = utils.RetryN(ctx, func() (bool, *trait.Error) {
			if err0 := e.Store.InsertJobLog(ctx, jobLogRecord); err0 != nil {
				e.Store.Log.Warnf("log job fail into store error: %s", err0.Error())
				return true, err0
			}
			return false, nil
		}, 3, 500*time.Millisecond)
	}

	// 建立目标任务中组件任务索引，避免后续错误删除
	target := job.Target.Components
	index := map[string]bool{}
	for _, c := range target {
		index[c.Component.Name] = true
	}

	sid := job.Current.SID
	s, err := e.getSystemContext(ctx, sid)
	if err != nil {
		l.Errorf("get system %d context error: %s", sid, err.Error())
		logErr(nil, err)
		setJobUninstallFail()
		return err
	}

	// get task list from work component table, rather then
	ts := []*trait.ComponentInstance{}
	indexMust := map[string]bool{}

	// 获取当前系统中任务项,避免遗漏.在后续代码中为目标任务中组件去重
	f := trait.WorkCompFilter{
		ListParam: trait.ListParam{
			Offset: 0,
			Limit:  100,
		},
		Aname: job.Current.AName,
		Sid:   sid,
	}

	// 获取当前系统组件项，以此为高优先级组件任务
	for {
		nodes, err := e.ListWorkComponentIns(ctx, f)
		if err != nil {
			return err
		}
		for _, n := range nodes {
			if ok := indexMust[n.Component.Name]; !ok {
				// uninstall task no need config in current impl, so set empty config
				ts = append(ts, &trait.ComponentInstance{
					ComponentInstanceMeta: *n,
				})
			}
		}
		if len(nodes) < f.Limit {
			break
		}
		f.Offset += len(nodes)
	}

	// 附加旧任务待清理项作为补充完善，优先级次之
	for _, n := range mustClean {
		if !indexMust[n.Component.Name] {
			ts = append(ts, n)
			indexMust[n.Component.Name] = true
		}
	}

	// 待清理组件与预期任务中组件进行比较去重
	deleteList := []*trait.ComponentInstance{}
	for _, c := range ts {
		if _, ok := index[c.Component.Name]; !ok {
			deleteList = append(deleteList, c)
		}
	}

	for _, t := range deleteList {
		cins := t
		e.Store.Log.Tracef(
			"uninstall component %s with id %d from system %d",
			cins.Component.Name, cins.CID, cins.System.SID)
		cins, err := e.Store.GetComponentIns(ctx, cins.CID)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			continue
		} else if err != nil {
			l.Errorf("get uninstall component instance error: %s", err.Error())
			logErr(nil, err)
			setJobUninstallFail()
			return err
		}

		// 组件迁移,所有权不同,不进行清理
		if cins.APPName != job.Target.Application.AName {
			continue
		}
		// 目前卸载无需其他组件定义信息，否则需要通过实例的acid获取对应组件定义
		meta := &trait.ComponentMeta{
			ComponentNode: cins.Component,
			Spec:          json.RawMessage("null"),
		}
		e.Store.Log.Tracef(
			"create uninstall taks for component %s  with type %s",
			cins.Component.Name, cins.Component.ComponentDefineType)
		ts, err := task.NewTask(meta, s)
		if err != nil {
			l.Errorf(
				"create uninstall taks for component %s  with type %s error: %s",
				cins.Component.Name, cins.Component.ComponentDefineType, err.Error(),
			)
			setJobUninstallFail()
			return err
		}

		// TODO check componentIns depenced edge

		if err := ts.SetComponentIns(cins); err != nil {
			com := ts.Component()
			l.Errorf("set component task config %s:%s error: %s", com.Name, com.Version, err.Error())
			logErr(cins, err)
			setJobUninstallFail()
			return err
		}
		e.Store.Log.Infof("uninstall component %s from system %d", cins.Component.Name, cins.System.SID)
		ts.WithLog(l)
		if err := ts.Uninstall(ctx); err != nil {
			if trait.IsInternalError(err, trait.ECBaseNode) {
				continue
			}
			l.Errorf("uninstall component %s in system [%s] error: %s", cins.Component.Name, s.NameSpace, err.Error())
			logErr(cins, err)
			setJobUninstallFail()
			return err
		}
		if err := e.removeWorkComponentIns(ctx, cins); err != nil {
			l.Errorf("remove work compomnent instance error: %s", err.Error())
			logErr(cins, err)
			setJobUninstallFail()
			return err
		}
	}

	jc.Lock()
	job.Target.Status = trait.AppUpgradeParentComponentStatus
	jc.Unlock()

	if err := e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, e.id, job.Target.StartTime, -1); err != nil {
		l.Errorf("change application status into deleted error: %s", err.Error())
		logErr(nil, err)
		return err
	}
	return nil
}

func (e *Executor) removeWorkComponentIns(ctx context.Context, ins *trait.ComponentInstance) *trait.Error {
	curCID := ins.CID
	tx, err := e.Store.Begin(ctx)
	if err != nil {
		e.Store.Log.Errorf("start transaction errorr when upgrade component in store: %s", err.Error())
		return err
	}

	// component must not  remove when component be depended, so need to delete EdgeTo
	if err := tx.DeleteEdgeFrom(ctx, curCID); err != nil {
		e.Store.Log.Errorf("layoff component edge clean error:%s", err.Error())
		e.rollbackWithLog(tx)()
		return err
	}

	if err := tx.LayoffComponentIns(ctx, curCID); err != nil {
		e.Store.Log.Errorf("layoff deleted component error:%s", err.Error())
		e.rollbackWithLog(tx)()
		return err
	}
	return tx.Commit()
}

func (e *Executor) newJobLogger() *log.TaskLogger {
	return log.NewTaskLogger(e.Log, _taskLogLevel, _taskLogCacheLen)
}

func (e *Executor) setWorkComponentIns(ctx context.Context, com *graph.ComponentNode, status int) *trait.Error {
	target := com.ComponentIns()
	tarCID := target.CID
	curCID := -1
	sid := com.ComponentIns().System.SID
	cur, err := e.Store.GetWorkComponentIns(ctx, target.System.SID, target.Component)
	if err == nil {
		curCID = cur.CID
	} else if trait.IsInternalError(err, trait.ErrNotFound) {
		err = nil
	} else {
		return err
	}

	tx, err := e.Store.Begin(ctx)
	if err != nil {
		e.Store.Log.Errorf("start transaction errorr when upgrade component in store: %s", err.Error())
		return err
	}

	if err0 := tx.UpdateComponentInsStatus(
		ctx, target.CID, status, target.Revission,
		target.StartTime, target.EndTime); err0 != nil &&
		!trait.IsInternalError(err, trait.ErrComponentInstanceRevission) {
		// 组件锁控制下,revision错误不存在,返回该错误只是因为上一次提交已成功,但未获得返回导致的重试,因此可以忽略.
		e.Store.Log.Errorf("please check db store, store component sucess status error:%s", err0.Error())
		defer e.rollbackWithLog(tx)()
		return err0
	}

	if curCID >= 0 {
		if err := tx.ChangeEdgeto(ctx, curCID, tarCID); err != nil {
			e.Store.Log.Errorf("change edge to target component instance %s error: %s", target.Component.Name, err.Error())
			e.rollbackWithLog(tx)()
			return err
		}

		if err := tx.ChangeEdgeFrom(ctx, curCID, tarCID); err != nil {
			e.Store.Log.Errorf("layoff component %s edge clean error:%s", target.Component.Name, err.Error())
			e.rollbackWithLog(tx)()
			return err
		}
		if err := tx.LayoffComponentIns(ctx, curCID); err != nil {
			e.Store.Log.Errorf("layoff component %s, error:%s", target.Component.Name, err.Error())
			e.rollbackWithLog(tx)()
			return err
		}
	}

	for _, child := range com.Children {
		// children edge has 4 type:
		// 1. need add/update outer children.
		// 2. need add/update inner children.
		// 3. need delete inner child which will be deleted, these will be delete when delete stage
		// 4. need add inner children which is new in this version
		to := child.ComponentIns().CID
		e.Store.Log.Tracef("add component %s edge (%d, %d)", target.Component.Name, tarCID, to)
		if child.Component().ComponentDefineType == component.ComponentBaseType {
			// outer children has been change by ChangeEdgeFrom
			// need add new outer children
			if err := tx.AddOuterChildEdge(ctx, tarCID, sid, *child.Component()); err != nil {
				e.Store.Log.Errorf("add component outer child edge (%d, %d) error:%s", tarCID, to, err.Error())
				return err
			}
			continue
		}
		// The inner children need to be updated has been change by children's ChangeEdgeTo and ChangeEdgeFrom.
		// So this only need to add edge which is new
		if err := tx.AddEdge(ctx, tarCID, to); err != nil {
			e.Store.Log.Errorf("add component inner child edge (%d, %d) error:%s", tarCID, to, err.Error())
			e.rollbackWithLog(tx)()
			return err
		}
	}

	e.Store.Log.Tracef("set work component %s, id %d", target.Component.Name, tarCID)
	if err := tx.WorkComponentIns(ctx, target); err != nil {
		e.Store.Log.Errorf("set component instance %s into work table error: %s", target.Component.Name, err.Error())
		e.rollbackWithLog(tx)()
		return err
	}

	// 延后revision调整,因为addEdge可能导致错误重试,从而使得revision调整失效
	target.Revission++
	err = tx.Commit()
	if err != nil {
		e.Store.Log.Errorf("set work instance transaction commit error: %s", err.Error())
	}
	return err
}
