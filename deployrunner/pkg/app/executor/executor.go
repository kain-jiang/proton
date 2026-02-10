package executor

import (
	"context"
	"fmt"

	"taskrunner/pkg/app"
	"taskrunner/pkg/cluster"
	"taskrunner/pkg/helm"
	"taskrunner/trait"

	pstore "taskrunner/pkg/store/proton"

	"helm.sh/helm/v3/pkg/time"
	"k8s.io/client-go/kubernetes"
)

// Executor execute jobRecord
type Executor struct {
	// lock     sync.Locker
	// ctx is run routine context, it will be set when Run(ctx)
	ctx context.Context
	*app.Store
	queue    freeWorkerQueue
	parallel int
	// TODO group helm client by system
	helmCli   helm.Client
	imageRepo cluster.ImageRepo
	id        int
	kcli      kubernetes.Interface
	pcli      *pstore.ProtonClient
}

// NewExecutor return a new RunnerEngine
func NewExecutor(s *app.Store, parrellel int, hcli helm.Client, imageRepo cluster.ImageRepo, id int, kcli kubernetes.Interface, pcli *pstore.ProtonClient) *Executor {
	if parrellel == 0 {
		parrellel = 1
	}
	return &Executor{
		ctx:       context.Background(),
		Store:     s,
		parallel:  parrellel,
		queue:     newFreeWorkerQueue(parrellel),
		helmCli:   hcli,
		imageRepo: imageRepo,
		id:        id,
		kcli:      kcli,
		pcli:      pcli,
	}
}

func (e *Executor) rollbackWithLog(tx trait.Transaction) func() {
	return func() {
		if err := tx.Rollback(); err != nil {
			e.Log.Errorf("rollback error: %s", err.Error())
		}
	}
}

func (e *Executor) CreateAndSetJobWithConfig(ctx context.Context, ins *trait.ApplicationInstance) (int, *trait.Error) {
	// TODO
	sid := ins.SID
	aid := ins.AID
	tx, err := e.Store.Begin(ctx)
	if err != nil {
		e.Log.Errorf("start a trasaction error, please contact env maintainer: %s", err.Error())
		return -1, err
	}

	j, err := e.Store.GetApplicationJobSnapshot(ctx, tx, aid, sid)
	if err != nil {
		defer e.rollbackWithLog(tx)()
		return -1, err
	}
	e.Store.MergeJobConfig(j, ins)
	if err := j.Target.Validate(); err != nil {
		defer e.rollbackWithLog(tx)()
		return -1, err
	}

	err = e.CheckDepComponents(ctx, &j.Target.Application, sid)
	if err != nil {
		defer e.rollbackWithLog(tx)()
		return -1, err
	}

	j.Target.Status = trait.AppConfirmedStatus
	// start job
	otype := ins.OType
	if otype != trait.JobDeleteOType {
		// 创建更新任务不允许设置为删除类型
		if j.Current == nil {
			j.Target.OType = trait.JobInstallOType
		} else {
			j.Target.OType = otype
		}
	}
	jid, err := tx.InsertJobRecord(ctx, j)
	if err != nil {
		e.Log.Errorf("create job record error: %s", err.Error())
		e.rollbackWithLog(tx)()
		return -1, err
	}
	j.ID = jid

	if err = tx.Commit(); err != nil {
		e.Log.Errorf("commit transaction error: %s", err.Error())
		return -1, err
	}
	return jid, err
}

// CreateAndStartJobWithConfig create and start the job
func (e *Executor) CreateAndStartJobWithConfig(ctx context.Context, ins *trait.ApplicationInstance) (int, *trait.Error) {
	jid, err := e.CreateAndSetJobWithConfig(ctx, ins)
	if err != nil {
		return -1, err
	}
	err = e.StartJob(ctx, jid)
	return jid, err
}

func (e *Executor) CreateDeleteJobAnStart(ctx context.Context, aid, sid int) (int, *trait.Error) {
	jid, err := e.NewJobRecordType(ctx, aid, sid, trait.JobDeleteOType, trait.AppConfirmedStatus)
	if err != nil {
		return -1, err
	}
	err = e.StartJob(ctx, jid)
	return jid, err
}

// CancelJob cancel job. It will block by start job
func (e *Executor) CancelJob(ctx context.Context, jid int) *trait.Error {
	jc := e.queue.jobControlIndex.IndexAndLock(jid)
	defer jc.Unlock()
	if jc.ctx != nil {
		// control by other routine, inform the routine with cancel function
		/*
			hold the lock until update status.
			the job's finish status is setted by run routine
			if don't hold the lock until update status, this update result may overwrite the finish status writed by run routine.
			the  locker may block the run routine.
		*/
		// jc.Lock()
		job := jc.job
		curStatus := job.Target.Status
		switch curStatus {
		case trait.AppWaitingStatus, trait.AppDoingStatus, trait.AppConfirmedStatus, trait.AppinitStatus, trait.AppUpgradeParentComponentStatus:
			job.Target.Status = trait.AppStopingStatus
		case trait.AppDeleteingOldComponentStatus, trait.AppUpdatedComponentStatus:
			err := fmt.Errorf("job shouldn't stop after status updated, current status is %d", curStatus)
			return &trait.Error{
				Err:      err,
				Internal: trait.ErrJobCantStop,
				Detail:   err.Error(),
			}
		case trait.AppSucessStatus, trait.AppFailStatus, trait.AppStopedStatus, trait.AppFailMissStatus, trait.AppFailUninstallStatus, trait.AppUpgradeParentComponentFailStatus:
			// has been done, don't do anything
			return nil
		default:
		}

		// TODO retry when error
		if err := e.Store.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, e.id, job.Target.StartTime, job.Target.EndTime); err != nil {
			jc.job.Target.Status = curStatus
			e.Store.Log.Errorf("change job status into waiting job eror: %s", err.Error())
			return err
		}

		// cancel in finaly step can rollback when db error
		jc.cancel()
	} else {
		defer e.queue.jobControlIndex.removeJobControl(jc)
		tx, err := e.Store.Begin(ctx)
		if err != nil {
			e.Store.Log.Errorf("start store transaction error: %s", err.Error())
			return err
		}

		jb, err := tx.GetJobRecord(ctx, jid)
		if err != nil {
			defer e.rollbackWithLog(tx)()
			e.Log.Errorf("get job %d info error: %s", jid, err.Error())
			return err
		}
		jc.job.Target = jb.Target
		if jc.job.Target.Onwer != 0 && jc.job.Target.Onwer != e.id {
			defer e.rollbackWithLog(tx)()
			return &trait.Error{
				Internal: trait.ErrJobOwnerError,
				Err:      fmt.Errorf("the job owner cross"),
				Detail:   fmt.Sprintf("job owner: %d, executor id: %d", jc.job.Target.Onwer, e.id),
			}
		}

		switch jc.job.Target.Status {
		case trait.AppSucessStatus, trait.AppFailStatus, trait.AppStopedStatus, trait.AppFailMissStatus, trait.AppFailUninstallStatus, trait.AppUpgradeParentComponentFailStatus:
			// has been done, don't do anything
			return tx.Rollback()
		}

		jb.Target.EndTime = int(time.Now().Unix())

		// 事务内清理锁，避免处理遗漏
		job := jc.job
		if err := tx.UnlockJobComponent(ctx, jc.job.ID); err != nil {
			e.Log.Errorf("UnlockJobComponent %d error: %s", jc.job.ID, err.Error())
			e.rollbackWithLog(tx)()
			return err
		}

		if err := tx.UnlockApp(ctx, job.Target.SID, job.ID, job.Target.AName); err != nil {
			e.Store.Log.Errorf("retry: unlock application: %s, jid: %d sid: %d, error: %s",
				job.Target.AName, job.ID, job.Target.SID, err.Error())
			e.rollbackWithLog(tx)()
			return err
		}

		// TODO retry when error
		if err := tx.UpdateAPPInsStatus(ctx, jid, trait.AppStopedStatus, e.id, jb.Target.StartTime, jb.Target.EndTime); err != nil {
			e.rollbackWithLog(tx)()
			e.Store.Log.Errorf("start waiting job fail when change job status: %s", err.Error())
			return err
		}
		return tx.Commit()
	}

	return nil
}

// StartJob start job. It will block by cancel
func (e *Executor) StartJob(ctx context.Context, jid int) *trait.Error {
	jc := e.queue.jobControlIndex.IndexAndLock(jid)
	defer jc.Unlock()

	if jc.ctx != nil {
		// the job has been start
		return &trait.Error{
			Internal: trait.ErrJobExecuting,
			Err:      fmt.Errorf("job context is hold, no need start again"),
			Detail:   jid,
		}
	}

	ctx0, cancel := trait.WithCancelCauesContext(e.ctx)
	jc.cancel = cancelJobFunc(cancel, jid)
	jc.ctx = ctx0

	qid := e.queue.hole(jc)
	if qid == -1 {
		e.Store.Log.Errorf("there isn't available worker, retry later")
		e.queue.jobControlIndex.removeJobControl(jc)
		return &trait.Error{
			Internal: app.ErrNoAvailableWorker,
			Err:      fmt.Errorf("no available worker"),
			Detail:   e.parallel,
		}
	}

	freeHole := func() {
		e.queue.freeNode(qid)
	}

	tx, err := e.Store.Begin(ctx)
	if err != nil {
		e.Store.Log.Errorf("start store transaction error: %s", err.Error())
		freeHole()
		return err
	}
	job, err := tx.GetJobRecord(ctx, jid)
	if err != nil {
		e.Store.Log.Errorf("get the job error: %s", err.Error())
		freeHole()
		e.rollbackWithLog(tx)()
		return err
	}

	if job.Target.Onwer != 0 && job.Target.Onwer != e.id {
		freeHole()
		e.rollbackWithLog(tx)()
		return &trait.Error{
			Internal: trait.ErrJobOwnerError,
			Err:      fmt.Errorf("the job owner cross"),
			Detail:   fmt.Sprintf("job owner: %d, executor id: %d", job.Target.Onwer, e.id),
		}
	}

	switch job.Target.Status {
	// case trait.AppWaitingStatus, trait.AppDoingStatus:
	// 	e.store.log.Infof("job %d is executing, couldn't start again", job.ID)
	// 	freeHole()
	// 	tx.Rollback(ctx)
	// 	return trait.ErrJobExecuting
	case trait.AppinitStatus:
		e.Store.Log.Infof("job %d isn't confirmed, please comfirm config first", job.ID)
		freeHole()
		e.rollbackWithLog(tx)()
		return &trait.Error{
			Internal: trait.ErrConfigNotComfirm,
			Err:      fmt.Errorf("job must confirmed before execute"),
		}
	}

	job.Target.Status = trait.AppWaitingStatus
	job.Target.StartTime = int(time.Now().Unix())
	job.Target.EndTime = -1
	if err := tx.UpdateAPPInsStatus(ctx, job.Target.ID, job.Target.Status, e.id, job.Target.StartTime, job.Target.EndTime); err != nil {
		e.Store.Log.Errorf("start inited job fail when change job status: %s", err.Error())
		e.rollbackWithLog(tx)()
		freeHole()
		return err
	}

	jc.job = &job
	if err := tx.Commit(); err != nil {
		e.Store.Log.Errorf("commit start job  %d error: %s", job.ID, err.Error())
		return err
	}

	e.queue.push(qid)
	e.Store.Log.Debugf("job target components length: %d", len(job.Target.Components))
	e.Store.Log.Debugf("job %d enqueue", job.ID)
	return nil
}

func (e *Executor) getSystemContext(ctx context.Context, sid int) (*cluster.SystemContext, *trait.Error) {
	// TODO
	system, err := e.Store.GetSystemInfo(ctx, sid)
	if err != nil {
		return nil, err
	}

	return &cluster.SystemContext{
		ImageRepo: e.imageRepo,
		System:    *system,
		HelmManagerInterface: cluster.HelmManagerInterface{
			HelmRepo:   e.Store.HelmRepo,
			HelmClient: e.helmCli,
		},
		Kcli: e.kcli,
	}, nil
}
