package executor

import (
	"context"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	"taskrunner/test"
	"taskrunner/trait"
)

var needPProf = os.Getenv("debug_pprof")

// ExecuteJob execute a job in waiting
func (e *Executor) ExecuteJob(ctx context.Context) *trait.Error {
	id, j := e.queue.pop(ctx)
	if id == -1 {
		e.Store.Log.Info("engine context dead, exit")
		return &trait.Error{
			Internal: trait.ECExit,
			Err:      context.Canceled,
			Detail:   "engine close sync queue, eixt",
		}
	}

	e.Store.Log.Infof("job %d start", j.job.ID)
	err := e.executeJob(ctx, j)
	e.queue.freeNode(id)
	e.Store.Log.Infof("job %d end", j.job.ID)
	return err
}

// Run block current routine, wait job enqueue then execute it
func (e *Executor) Run(ctx context.Context) {
	if needPProf != "" {
		fout, err := test.StartCPUPProf("")
		if err != nil {
			e.Log.Error(err)
			return
		}
		defer fout.Close()
		defer pprof.StopCPUProfile()

	}

	e.ctx = ctx

	wg := &sync.WaitGroup{}
	wg.Add(e.parallel)
	for i := 0; i < e.parallel; i++ {
		go func() {
			defer wg.Done()
			for !trait.IsInternalError(e.ExecuteJob(ctx), trait.ECExit) {
			}
		}()
	}
	wg.Wait()

	if needPProf != "" {
		foutm, err := test.StartMemoryPProf("")
		if err != nil {
			e.Log.Error(err)
			return
		}
		defer foutm.Close()
	}
}

func (e *Executor) snapshotInterruptJob(ctx context.Context) ([]int, *trait.Error) {
	jids := []int{}
	offset := 0
	for {
		jbs, err := e.Store.ListJobRecord(ctx, &trait.AppInsFilter{
			Sid:    -1,
			Limit:  10,
			Offset: offset,
			Status: trait.JobDoingStauts,
		})
		if trait.IsInternalError(err, trait.ErrNotFound) || len(jbs) == 0 {
			break
		}
		if err != nil {
			e.Store.Log.Errorf("get interrupt job error")
			return nil, err
		}
		l := len(jbs)
		offset += l

		for _, jb := range jbs {
			jc := e.queue.jobControlIndex.IndexAndLock(jb.ID)
			if jc.ctx != nil {
				// skip the job has been start
				jc.Unlock()
				continue
			}
			e.queue.jobControlIndex.removeJobControl(jc)
			jc.Unlock()
			jids = append(jids, jc.job.ID)
		}
	}
	return jids, nil
}

func (e *Executor) reEnqueueInterruptJob(ctx context.Context, jids []int) {
	interval := 3 * time.Second
	for _, jid := range jids {
		jc := e.queue.jobControlIndex.IndexAndLock(jid)
		if jc.ctx != nil {
			e.Store.Log.Infof("job %d has been executed outside recovery routine, ignore it", jc.job.ID)
			jc.Unlock()
			continue
		}
		ctx0, cancel := trait.WithCancelCauesContext(e.ctx)
		jc.cancel = cancelJobFunc(cancel, jid)
		jc.ctx = ctx0

		qid := -1
		for qid == -1 {
			qid = e.queue.hole(jc)
			if qid == -1 {
				e.Store.Log.Errorf("there isn't available worker, recoery job %d retry later", jid)
				// avoid other job operator block in acquire lock, the job control shuold release the lock
				jc.Unlock()
				timer := time.NewTimer(interval)
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					jc.Lock()
				}
			}
		}

		job := jc.job

		for job.Target == nil {
			j, err := e.Store.GetJobRecord(ctx, jid)
			if trait.IsInternalError(err, trait.ErrNotFound) {
				e.queue.freeNode(qid)
				jc.Unlock()
				e.Store.Log.Infof("the job %d has been removed, ignore it in  recovery routine", jid)
				return
			}
			if err != nil {
				jc.Unlock()
				e.Store.Log.Errorf("recover routine get the job error: %s", err.Error())
				timer := time.NewTimer(interval)
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					jc.Lock()
				}
			} else {
				job = &j
				break
			}
		}

		if !hasInt(trait.JobDoingStauts, job.Target.Status) {
			// ignore sucess job
			jc.Unlock()
			e.queue.freeNode(qid)
		} else {
			jc.job = job
			e.Store.Log.Infof("job %d enqueue from recovery routine,", jc.job.ID)
			e.queue.push(qid)
			jc.Unlock()
		}

	}
}

// Recover recover the interrupt job.
// It will enqueue the job then execute by Run routine.
// it's context must the context for Run
// Warn: this func must call before Run
// The difference between Rcover and StartJob is that it does not start non interrupt job
func (e *Executor) Recover(ctx context.Context) (*sync.WaitGroup, *trait.Error) {
	jids, err := e.snapshotInterruptJob(ctx)
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	e.ctx = ctx
	go func() {
		defer wg.Done()
		e.reEnqueueInterruptJob(ctx, jids)
	}()
	return wg, nil
}

func hasInt(array []int, want int) bool {
	for _, j := range array {
		if want == j {
			return true
		}
	}
	return false
}
