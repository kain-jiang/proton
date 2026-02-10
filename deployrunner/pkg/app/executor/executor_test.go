package executor

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"taskrunner/pkg/app"
	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component"
	"taskrunner/pkg/graph"
	"taskrunner/pkg/helm"
	"taskrunner/test"
	"taskrunner/test/mock"
	"taskrunner/trait"

	"github.com/mohae/deepcopy"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes/fake"
)

var getTestAppliationBytes = mock.GetTestAppliationBytes

func testStoreInstance(t *testing.T) app.Store {
	repo := mock.HelmRepoMock{
		RepoName: "test",
		Chart:    map[string][]byte{},
		WantErr:  nil,
	}
	s := app.NewStore(logrus.New(), &mock.DbStoreFaker{
		ErrMap: make(map[string]*trait.Error),
	}, helm.NewHelmIndexRepo(&repo))

	s.Log.SetReportCaller(true)
	return *s
}

func TestFressWorkerQueue(t *testing.T) {
	tt := test.TestingT{T: t}
	num := 10
	q := newFreeWorkerQueue(num)
	wg := &sync.WaitGroup{}
	wg.Add(num)
	var err error
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			jc := &jobControl{
				job: &trait.JobRecord{
					ID: 0,
				},
			}
			id := q.hole(jc)
			jc.job.ID = id
			q.push(id)

			id, jc = q.pop(context.Background())
			if jc.job.ID != id {
				err = fmt.Errorf("%d!=%d", id, jc.job.ID)
			}
		}()
	}
	wg.Wait()
	tt.AssertNil(err)

	ctx, cancel := trait.WithCancelCauesContext(context.Background())
	cancel(&trait.Error{
		Internal: trait.ECJobCancel,
		Err:      context.Canceled,
	})
	q.pop(ctx)

	id := q.hole(&jobControl{})
	if id != -1 {
		t.Fatal(id)
	}

	q.freeNode(1)
	id = q.hole(&jobControl{job: &trait.JobRecord{}})
	if id != 1 {
		t.Fatal(id)
	}
}

func TestJobsINdex(t *testing.T) {
	ji := newJobsIndex()
	job := &trait.JobRecord{
		ID: 10,
	}

	jc := ji.IndexAndLock(job.ID)
	// jc.job = job
	if jc == nil {
		t.FailNow()
	}
	ctx, cancel := trait.WithCancelCauesContext(context.Background())
	jc.ctx = ctx
	jc.cancel = cancelJobFunc(cancel, jc.job.ID)

	jc.Unlock()

	jc0 := ji.IndexAndLock(job.ID)
	if jc != jc0 {
		t.Fatal(jc0)
	}

	ji.removeJobControl(jc0)
	jc0 = ji.IndexOrCreate(job.ID)
	if jc0 == jc {
		t.Fatal(jc0)
	}
}

func TestStartJob(t *testing.T) {
	s := testStoreInstance(t)
	tt := test.TestingT{T: t}
	s.Log.SetLevel(logrus.FatalLevel)

	hcli := &mock.HelmCliMock{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, cluster.ImageRepo{}, -1, kcli, nil)
	bs := getTestAppliationBytes(t)
	ctx := context.Background()
	db := s.Store.(*mock.DbStoreFaker)

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)

	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	jid, err := e.Store.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)
	jins, err := db.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	ains := db.AppInsCache[jins.Target.ID]
	if ains != jins.Target {
		t.Fatal(ains, jins.Target)
	}
	err = e.Store.SetJobConfig(ctx, jid, jins.Target)
	tt.AssertNil(err)

	err0 := &trait.Error{
		Internal: trait.ECNULL,
		Err:      fmt.Errorf("beginErr"),
	}
	db.ErrMap["Begin"] = err0
	err = e.StartJob(ctx, jid)
	tt.Assert(err0, err)
	db.ErrMap["Begin"] = nil

	err0.Err = fmt.Errorf("GetJobRecordError")
	db.ErrMap["GetJobRecord"] = err0
	err = e.StartJob(ctx, jid)
	db.ErrMap["GetJobRecord"] = nil
	tt.Assert(err0, err)

	err0.Err = fmt.Errorf("UpdateAPPInsStatus")
	db.ErrMap["UpdateAPPInsStatus"] = err0
	err = e.StartJob(ctx, jid)
	db.ErrMap["UpdateAPPInsStatus"] = nil
	tt.Assert(err0, err)

	// err = e.StartJob(ctx, jid)
	// tt.Assert(trait.ErrJobExecuting, err)

	err = db.UpdateAPPInsStatus(ctx, jins.Target.ID, trait.AppinitStatus, e.id, -1, -1)
	tt.AssertNil(err)
	err = e.StartJob(ctx, jid)
	tt.AssertError(trait.ErrConfigNotComfirm, err)

	err = db.UpdateAPPInsStatus(ctx, jins.Target.ID, trait.AppConfirmedStatus, e.id, -1, -1)
	tt.AssertNil(err)

	err0.Err = fmt.Errorf("commit error mock")
	db.ErrMap["Commit"] = err0
	err = e.StartJob(ctx, jid)
	tt.Assert(err0, err)

	db.ErrMap["Commit"] = nil
	err = e.StartJob(ctx, jid)
	tt.AssertError(trait.ErrJobExecuting, err)

	err = e.StartJob(ctx, jid+1)
	tt.AssertError(app.ErrNoAvailableWorker, err)
}

func TestCancelJob(t *testing.T) {
	tt := test.TestingT{T: t}
	// init data
	s := testStoreInstance(t)
	s.Log.SetLevel(logrus.FatalLevel)
	hcli := &mock.HelmCliMock{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, cluster.ImageRepo{}, 0, kcli, nil)
	bs := getTestAppliationBytes(t)
	ctx := context.Background()
	db := s.Store.(*mock.DbStoreFaker)

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)

	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	jid, err := e.Store.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)
	jins, err := db.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	ains := db.AppInsCache[jins.Target.ID]
	if ains != jins.Target {
		t.Fatal(ains, jins.Target)
	}
	err = e.Store.SetJobConfig(ctx, jid, jins.Target)
	tt.AssertNil(err)

	// init data end

	// only change store data status
	wantErr := &trait.Error{
		Internal: trait.ECNULL,
		Err:      fmt.Errorf("UpdateAPPInsStatus mock error"),
	}
	db.ErrMap["UpdateAPPInsStatus"] = wantErr
	err = e.CancelJob(ctx, jins.ID)
	tt.Assert(wantErr, err)
	db.ErrMap["UpdateAPPInsStatus"] = nil
	err = e.CancelJob(ctx, jins.ID)
	tt.AssertNil(err)
	jins, err = db.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	if jins.Target.Status != trait.AppStopedStatus {
		t.Fatal(jins.Target.Status)
	}

	// cancel the doing job
	err = e.StartJob(ctx, jid)
	tt.AssertNil(err)
	err = e.CancelJob(ctx, jins.ID)
	tt.AssertNil(err)
	jins, err = db.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	if jins.Target.Status != trait.AppStopingStatus {
		t.Fatal(jins.Target.Status)
	}

	jc := e.queue.jobControlIndex.IndexAndLock(jid)
	jc.Unlock()
	jc.job.Target.Status = trait.AppUpdatedComponentStatus
	err = e.CancelJob(ctx, jins.ID)
	tt.AssertError(trait.ErrJobCantStop, err)

	jc.job.Target.Status = trait.AppDoingStatus
	db.ErrMap["UpdateAPPInsStatus"] = wantErr
	err = e.CancelJob(ctx, jins.ID)
	tt.Assert(wantErr, err)
	db.ErrMap["UpdateAPPInsStatus"] = nil
	err = e.CancelJob(ctx, jins.ID)
	tt.AssertNil(err)

	jins, err = db.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	if jins.Target.Status != trait.AppStopingStatus {
		t.Fatal(jins.Target.Status)
	}
}

func TestExecuteJob(t *testing.T) {
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	s.Log.SetLevel(logrus.ErrorLevel)
	hcli := &mock.HelmCliMock{}
	irepo := cluster.ImageRepo{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, irepo, -1, kcli, nil)

	bs := getTestAppliationBytes(t)
	ctx := context.Background()
	db := s.Store.(*mock.DbStoreFaker)

	_, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)

	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)

	jid, err := e.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)
	err = e.SetJobConfig(ctx, jid, &trait.ApplicationInstance{})
	tt.AssertNil(err)
	err = e.StartJob(ctx, jid)
	tt.AssertNil(err)
	err = e.ExecuteJob(ctx)
	tt.AssertNil(err)

	app, err := db.GetAPP(ctx, aid)
	tt.AssertNil(err)
	ains, err := db.GetWorkAPPIns(ctx, app.AName, sid)
	tt.AssertNil(err)

	tt.Assert(ains.AID, aid)

	func() {
		// test error  with block
		err = e.StartJob(ctx, jid)
		tt.AssertNil(err)
		id, j := e.queue.pop(ctx)
		defer e.queue.freeNode(id)
		err := e.executeJob(ctx, j)
		tt.AssertNil(err)
		j.job.Target.Status = trait.AppStopingStatus

		err = e.executeJob(ctx, j)
		tt.AssertNil(err)

		err = e.executeJob(ctx, j)
		tt.AssertNil(err)
		tt.Assert(trait.AppStopedStatus, j.job.Target.Status)
	}()

	func() {
		_, err := e.CreateDeleteJobAnStart(ctx, aid, sid)
		tt.AssertNil(err)
		id, j := e.queue.pop(ctx)
		defer e.queue.freeNode(id)
		err = e.executeJob(ctx, j)
		tt.AssertNil(err)
	}()

	{
		_, err := e.CreateAndStartJobWithConfig(ctx, &trait.ApplicationInstance{
			ApplicationinstanceMeta: trait.ApplicationinstanceMeta{
				System: trait.System{
					SID: sid,
				},
			},
			Application: trait.Application{
				ApplicationMeta: trait.ApplicationMeta{
					AID: aid,
				},
			},
			Trait: trait.ApplicationTrait{
				UpgradeParent: true,
			},
		})
		tt.AssertNil(err)
		id, j := e.queue.pop(ctx)
		defer e.queue.freeNode(id)
		err = e.executeJob(ctx, j)
		tt.AssertNil(err)
	}
}

func TestExecuteTask(t *testing.T) {
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	s.Log.SetLevel(logrus.ErrorLevel)
	hcli := &mock.HelmCliMock{}
	irepo := cluster.ImageRepo{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, irepo, -1, kcli, nil)

	bs := getTestAppliationBytes(t)
	ctx := context.Background()
	db := s.Store.(*mock.DbStoreFaker)

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)

	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	jid, err := e.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)
	job, err := e.GetJobRecord(ctx, jid)
	tt.AssertNil(err)

	resetJob := func() {
		for _, c := range job.Target.Components {
			c.Status = trait.AppWaitingStatus
		}
	}
	jc := newJobControl(&job)
	ctx0, cancel := trait.WithCancelCauesContext(ctx)
	jc.cancel = cancelJobFunc(cancel, job.ID)
	jc.ctx = ctx0
	tryPlanDbErr := func(plan *graph.Plan, fn string) {
		err0 := &trait.Error{
			Err:      fmt.Errorf("%s mock error", fn),
			Internal: trait.ECNULL,
		}
		db.ErrMap[fn] = err0
		err := e.executeTask(ctx, jc, plan)
		tt.Assert(err0, err)
		db.ErrMap[fn] = nil
	}

	tryErr := func(err0 *trait.Error) {
		plan, err := e.NewPlan(ctx, &job)
		tt.AssertNil(err)
		err1 := e.executeTask(ctx, jc, plan)
		tt.Assert(err0, err1)
	}

	tryNilErr := func() {
		plan, err := e.NewPlan(ctx, &job)
		tt.AssertNil(err)
		err1 := e.executeTask(ctx, jc, plan)
		tt.AssertNil(err1)
	}

	tryDBErr := func(fn string) {
		resetJob()
		err0 := &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("%s mock error", fn),
		}
		db.ErrMap[fn] = err0
		tryErr(err0)
		db.ErrMap[fn] = nil
	}

	tryDBErr("LockComponent")
	tryDBErr("GetComponentIns")
	tryDBErr("GetWorkComponentIns")
	tryDBErr("UpdateComponentInsStatus")
	tryDBErr("Begin")

	err0 := &trait.Error{
		Internal: trait.ECNULL,
		Err:      fmt.Errorf("component install mock error"),
	}
	hcli.Err = err0

	resetJob()
	tryErr(hcli.Err)
	hcli.Err = nil
	resetJob()
	tryNilErr()

	db.ErrMap["UnlockComponent"] = err0
	resetJob()
	tryNilErr()

	db.ErrMap["UnlockComponent"] = nil

	tryNilErr()

	resetPlanForTopology := func() *graph.Plan {
		for i, c := range job.Target.Components {
			c.Revission++
			cc := deepcopy.Copy(c).(*trait.ComponentInstance)
			cc.Revission--
			job.Target.Components[i] = cc

		}
		resetJob()
		plan, err := e.NewPlan(ctx, &job)
		tt.AssertNil(err)
		return plan
	}

	externalNode := &trait.ComponentInstance{
		ComponentInstanceMeta: trait.ComponentInstanceMeta{
			System: job.Target.System,
			Component: trait.ComponentNode{
				Name: "testexternal",
			},
		},
	}

	ecid, err := db.InsertComponentIns(ctx, externalNode)
	tt.AssertNil(err)
	externalNode.CID = ecid
	err = e.WorkComponentIns(ctx, externalNode)
	tt.AssertNil(err)

	app, err := db.GetAPP(ctx, aid)
	tt.AssertNil(err)
	app.Graph = append(app.Graph, trait.Edge{
		From: app.Component[0].ComponentNode,
		To: trait.ComponentNode{
			Name: "testexternal",
		},
	})

	plan := resetPlanForTopology()
	tryPlanDbErr(plan, "GetWorkComponentIns")

	plan = resetPlanForTopology()
	// err = &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("")}
	db.ErrMap["GetWorkComponentIns"] = &trait.Error{Internal: trait.ErrNotFound, Err: fmt.Errorf("")}
	tt.AssertError(trait.ErrComponentNotFound, e.executeTask(ctx, jc, plan))
	db.ErrMap["GetWorkComponentIns"] = nil

	plan = resetPlanForTopology()
	tt.AssertNil(e.executeTask(ctx, jc, plan))
}

func TestDeleteTask(t *testing.T) {
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	hcli := &mock.HelmCliMock{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, cluster.ImageRepo{}, 0, kcli, nil)
	bs := getTestAppliationBytes(t)
	ctx := context.Background()

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)
	db := s.Store.(*mock.DbStoreFaker)
	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	ss, err := s.GetSystemInfo(ctx, sid)
	tt.AssertNil(err)
	a, err := s.GetAPP(ctx, aid)
	tt.AssertNil(err)
	ains, err := app.NewAPPIns(nil, *ss, *a)
	tt.AssertNil(err)
	id, err := e.Store.InsertAPPIns(ctx, ains)
	tt.AssertNil(err)
	ains.ID = id
	err = e.Store.WorkAppIns(ctx, ains)
	tt.AssertNil(err)

	jid, err := s.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)

	job, err := s.GetJobRecord(ctx, jid)
	tt.AssertNil(err)

	//nolint: fake test won't error
	_ = db.AddEdge(ctx, job.Current.Components[0].CID, job.Current.Components[1].CID)
	job.Target.Components[1].APPName = "change"

	job.Target.Components = []*trait.ComponentInstance{
		{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				Component: trait.ComponentNode{
					Name:                "python4",
					ComponentDefineType: component.ComponentHelmTaskType,
				},
			},
		},
	}

	job.Current.Components = []*trait.ComponentInstance{
		{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				Component: trait.ComponentNode{
					Name:                "python4",
					ComponentDefineType: component.ComponentHelmTaskType,
				},
			},
		},
		{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				Component: trait.ComponentNode{
					Name:                "testnotfound",
					ComponentDefineType: component.ComponentHelmTaskType,
				},
			},
		},
		{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				Component: trait.ComponentNode{
					Name:                "python3",
					ComponentDefineType: component.ComponentHelmTaskType,
				},
			},
		},
	}

	ctx0, cancle := trait.WithCancelCauesContext(ctx)
	defer cancle(nil)

	jc := &jobControl{
		Locker: &sync.Mutex{},
		cancel: cancelJobFunc(cancle, job.ID),
		ctx:    ctx0,
		job:    &job,
	}
	tryErr := func(fn string) {
		err0 := &trait.Error{
			Err: fmt.Errorf("%s mock error", fn),
		}

		db.ErrMap[fn] = err0
		err := e.deleteTask(ctx, jc, nil)
		db.ErrMap[fn] = nil
		tt.Assert(err0, err)
	}
	tryErr("UpdateAPPInsStatus")

	job.Current.SID = -1
	err = e.deleteTask(ctx, jc, nil)
	tt.AssertError(trait.ErrNotFound, err)
	job.Current.SID = sid

	// test with mock
	err0 := &trait.Error{
		Internal: trait.ECNULL,
		Err:      fmt.Errorf("test mock error"),
	}
	hcli.Err = err0
	err = e.deleteTask(ctx, jc, nil)
	tt.Assert(hcli.Err, err)
	hcli.Err = nil

	tryErr("ListWorkComponentIns")

	tryErr("GetComponentIns")
	tryErr("Begin")
	tryErr("DeleteEdgeFrom")
	tryErr("LayoffComponentIns")

	// test with mock end

	err = e.deleteTask(ctx, jc, nil)
	tt.AssertNil(err)
	c, err := s.GetWorkComponentIns(ctx, sid, trait.ComponentNode{
		Name: "python4",
	})
	tt.AssertNil(err)
	if c == nil || c.APPName != "change" {
		t.Fatal(c)
	}

	jc.job.Current = nil
	err = e.deleteTask(ctx, jc, nil)
	tt.AssertNil(err)
}

func TestEngineRun(t *testing.T) {
	// init data
	// t.SkipNow()
	// TODO: FAIL
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	s.Log.SetLevel(logrus.InfoLevel)
	hcli := &mock.HelmCliMock{}
	irepo := cluster.ImageRepo{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, irepo, -1, kcli, nil)

	bs := getTestAppliationBytes(t)
	ctx := context.Background()
	db := s.Store.(*mock.DbStoreFaker)

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)

	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)

	runJob := func() {
		ctx0, cancel := trait.WithTimeoutCauseContext(ctx, 500*time.Millisecond, &trait.Error{
			Internal: trait.ECExit,
			Err:      fmt.Errorf("testEngineRun"),
			Detail:   "",
		})
		defer cancel()
		e.ctx = ctx0
		jid, err := e.Store.NewJobRecord(ctx, aid, sid)
		tt.AssertNil(err)
		jins, err := db.GetJobRecord(ctx, jid)
		tt.AssertNil(err)
		ains := db.AppInsCache[jins.Target.ID]
		if ains != jins.Target {
			t.Fatal(ains, jins.Target)
		}
		err = e.Store.SetJobConfig(ctx, jid, jins.Target)
		tt.AssertNil(err)

		err = e.StartJob(ctx, jid)
		tt.AssertNil(err)
		// init data end

		e.Store.Log.Info("start main routine")

		// <-ctx0.Done()
		e.Run(ctx0)
		jins, err = db.GetJobRecord(ctx, jid)
		tt.AssertNil(err)
		if jins.Target.Status != trait.AppSucessStatus {
			t.Fatal(jins.Target.Status)
		}
	}

	runJob()
	runJob()
}

func TestEngineRecover(t *testing.T) {
	// init data
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	hcli := &mock.HelmCliMock{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, cluster.ImageRepo{}, -1, kcli, nil)
	bs := getTestAppliationBytes(t)
	ctx := context.Background()
	db := s.Store.(*mock.DbStoreFaker)

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)

	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	jid, err := e.Store.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)
	jins, err := db.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	ains := db.AppInsCache[jins.Target.ID]
	if ains != jins.Target {
		t.Fatal(ains, jins.Target)
	}
	err = e.Store.SetJobConfig(ctx, jid, jins.Target)
	tt.AssertNil(err)

	jins.Target.StartTime = int(time.Now().Unix())
	jins.Target.EndTime = -1

	err = e.Store.UpdateAPPInsStatus(ctx, ains.ID, trait.AppDoingStatus, e.id, jins.Target.StartTime, jins.Target.EndTime)
	tt.AssertNil(err)

	// init data end
	err = &trait.Error{
		Internal: trait.ECNULL,
		Err:      fmt.Errorf("ListJobRecordExecutingError"),
	}
	db.ErrMap["ListJobRecordExecuting"] = err
	_, err0 := e.snapshotInterruptJob(ctx)
	tt.Assert(err, err0)
	db.ErrMap["ListJobRecordExecuting"] = nil

	jc := e.queue.jobControlIndex.IndexAndLock(jid)
	jc.ctx = ctx
	jc.Unlock()

	jids, err0 := e.snapshotInterruptJob(ctx)
	tt.AssertNil(err0)
	if len(jids) != 0 {
		t.Fatal(jids)
	}
	e.queue.jobControlIndex.removeJobControl(jc)

	jids, err0 = e.snapshotInterruptJob(ctx)
	tt.AssertNil(err0)
	if len(jids) != 1 {
		t.Fatal(jids)
	}

	// onwership hole
	jc = e.queue.jobControlIndex.IndexAndLock(jid)
	jc.ctx = ctx
	jc.Unlock()
	e.reEnqueueInterruptJob(ctx, jids)
	qid := e.queue.hole(jc)
	if qid == -1 {
		t.Fatal(qid)
	}

	// no worker
	jc.ctx = nil
	ctx0, cancel := trait.WithTimeoutCauseContext(ctx, 500*time.Millisecond, nil)
	defer cancel()
	e.reEnqueueInterruptJob(ctx0, jids)
	jc = e.queue.jobControlIndex.IndexAndLock(jid)
	if jc.job.Target != nil {
		t.Fatal(jc)
	}
	e.queue.freeNode(qid)
	jc.Unlock()

	// db mock get jobrecord error
	err.Err = fmt.Errorf("GetJobRecord")
	db.ErrMap["GetJobRecord"] = err
	ctx0, cancel = trait.WithTimeoutCauseContext(ctx, 500*time.Millisecond, nil)
	defer cancel()
	e.reEnqueueInterruptJob(ctx0, jids)
	db.ErrMap["GetJobRecord"] = nil
	qid = e.queue.hole(jc)
	if qid != -1 {
		t.Fatal(qid)
	}
	e.queue.freeNode(0)

	err.Err = fmt.Errorf("GetJobRecord")
	err.Internal = trait.ErrNotFound
	db.ErrMap["GetJobRecord"] = err
	e.reEnqueueInterruptJob(ctx, jids)
	db.ErrMap["GetJobRecord"] = nil
	if jc.job.Target != nil {
		t.Fatal(jc.job)
	}

	wg, err := e.Recover(ctx)
	tt.AssertNil(err)
	wg.Wait()
	qid, jc = e.queue.pop(ctx)
	e.queue.freeNode(qid)

	jc.job.Target.Status = trait.AppSucessStatus
	e.reEnqueueInterruptJob(ctx, jids)
	select {
	case jid = <-e.queue.signal:
		t.Fatal(jid)
	default:
	}
}
