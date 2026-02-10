package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"taskrunner/pkg/app"
	"taskrunner/pkg/app/builder"
	"taskrunner/pkg/cluster"
	"taskrunner/pkg/graph"
	"taskrunner/pkg/graph/task"
	"taskrunner/test"
	"taskrunner/test/mock"
	"taskrunner/trait"

	"k8s.io/client-go/kubernetes/fake"
)

func TestNewPlan(t *testing.T) {
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, nil, cluster.ImageRepo{}, 0, kcli, nil)
	bs := getTestAppliationBytes(t)
	ctx := context.Background()

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)
	// db := s.Store.(*dbStoreFaker)
	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	jid, err := s.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)

	job, err := s.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	plan, err := e.NewPlan(ctx, &job)
	tt.AssertNil(err)
	defer plan.Close()

	job.Target.System.SID = 10
	_, err = e.NewPlan(ctx, &job)
	tt.AssertError(trait.ErrNotFound, err)
}

func TestGetExternalAttribute(t *testing.T) {
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, nil, cluster.ImageRepo{}, 0, kcli, nil)

	ctx := context.Background()
	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	mock := e.Store.Store.(*mock.DbStoreFaker)
	cins := &trait.ComponentInstance{
		ComponentInstanceMeta: trait.ComponentInstanceMeta{
			System: trait.System{
				SID: sid,
			},
			Component: trait.ComponentNode{
				Name:    "python4",
				Version: "0.1.1",
			},
		},
	}
	cid, err := mock.InsertComponentIns(ctx, cins)
	tt.AssertNil(err)
	cins.CID = cid

	err = e.Store.WorkComponentIns(ctx, cins)
	tt.AssertNil(err)
	ts := []*task.Base{
		{
			ComponentInsData: &trait.ComponentInstance{
				ComponentInstanceMeta: trait.ComponentInstanceMeta{
					Component: trait.ComponentNode{
						Name:    "python4",
						Version: "0.1.0",
					},
				},
			},
		},
	}
	err = e.gotExternalAttribute(ctx, sid, ts)
	tt.AssertNil(err)

	ts[0].ComponentInsData.Component.Version = "0.1.2"
	err = e.gotExternalAttribute(ctx, sid, ts)
	tt.AssertNil(err)

	ts[0].ComponentInsData.Component.Version = "0.2.1"
	err = e.gotExternalAttribute(ctx, sid, ts)
	tt.AssertError(trait.ErrComponentVersionLess, err)

	ts[0].ComponentInsData.Component.Name = "python123"
	err = e.gotExternalAttribute(ctx, sid, ts)
	if !trait.IsInternalError(err, trait.ErrComponentNotFound) {
		t.Fatal(err)
	}
}

func TestNewAPPIns(t *testing.T) {
	tt := test.TestingT{T: t}
	bs := getTestAppliationBytes(t)
	// ctx := context.Background()
	application, _, err := builder.ParseApplication(bytes.NewReader(bs))

	tt.AssertNil(err)
	ss := trait.System{}
	ins, err := app.NewAPPIns(nil, ss, application)
	tt.AssertNil(err)
	want := map[string]bool{
		"python3:0.1.0": false,
		"python4:0.1.0": false,
	}
	for _, c := range ins.Components {
		want[c.Component.Name+":"+c.Component.Version] = true
	}

	for k, v := range want {
		if !v {
			t.Fatal(k)
		}
	}
}

func TestSetWorkComponentIns(t *testing.T) {
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, nil, cluster.ImageRepo{}, 0, kcli, nil)
	ctx := context.Background()
	db := s.Store.(*mock.DbStoreFaker)
	sid, _ := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	com := &graph.ComponentNode{
		Task: &task.Base{
			ComponentInsData: &trait.ComponentInstance{
				ComponentInstanceMeta: trait.ComponentInstanceMeta{
					System: trait.System{
						SID: sid,
					},
					CID: 1,
					Component: trait.ComponentNode{
						Name:                "test",
						Version:             "0.1.0",
						ComponentDefineType: "base",
					},
				},
			},
		},
		Children: []*graph.ComponentNode{
			{
				Task: &task.Base{
					ComponentInsData: &trait.ComponentInstance{
						ComponentInstanceMeta: trait.ComponentInstanceMeta{
							System: trait.System{
								SID: sid,
							},
							CID: 2,
						},
					},
				},
			},
		},
	}
	cid, err := db.InsertComponentIns(ctx, com.ComponentIns())
	tt.AssertNil(err)
	com.ComponentIns().CID = cid
	tryErr := func(fn string) {
		err0 := &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("%s error mock", fn),
		}
		db.ErrMap[fn] = err0
		err = e.setWorkComponentIns(ctx, com, com.ComponentIns().Status)
		tt.Assert(err0, err)
		db.ErrMap[fn] = nil
	}

	tryErr("GetWorkComponentIns")
	tryErr("Begin")
	tryErr("WorkComponentIns")
	tryErr("AddEdge")
	err = e.setWorkComponentIns(ctx, com, com.ComponentIns().Status)
	tt.AssertNil(err)
	c, err := s.CountEdgeTo(ctx, 2)
	tt.AssertNil(err)
	tt.Assert(1, c)

	tryErr("ChangeEdgeto")
	tryErr("ChangeEdgeFrom")
	tryErr("LayoffComponentIns")
	err = e.setWorkComponentIns(ctx, com, com.ComponentIns().Status)
	tt.AssertNil(err)
	c, err = s.CountEdgeTo(ctx, 2)
	tt.AssertNil(err)
	tt.Assert(1, c)
}

func TestDebugAsMain(t *testing.T) {
	t.SkipNow()
	tt := test.TestingT{T: t}
	bs, err := os.ReadFile("")
	tt.AssertNil(err)
	s := testStoreInstance(t)
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, nil, cluster.ImageRepo{}, 0, kcli, nil)
	// ctx := context.Background()
	job := &trait.JobRecord{
		Target: &trait.ApplicationInstance{},
	}
	err = json.Unmarshal(bs, &job.Target)
	tt.AssertNil(err)
	plan, external, err := e.newJobs(job, cluster.SystemContext{})
	tt.AssertNil(err)
	tt.AssertNil(plan)
	tt.AssertNil(external)
}

func TestHeavyUpdateParent(t *testing.T) {
	tt := test.TestingT{T: t}
	s := testStoreInstance(t)
	hcli := &mock.HelmCliMock{}
	irepo := cluster.ImageRepo{}
	kcli := fake.NewSimpleClientset()
	e := NewExecutor(&s, 1, hcli, irepo, -1, kcli, nil)
	ctx := context.Background()
	appBytes := getTestAppliationBytes(t)
	aid, err := e.UploadApplicationPackage(ctx, bytes.NewReader(appBytes))
	tt.AssertNil(err)

	sid, err := e.InsertSystemInfo(ctx, trait.System{})
	tt.AssertNil(err)

	jid, err := e.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)
	err = e.SetJobConfig(ctx, jid, &trait.ApplicationInstance{})
	tt.AssertNil(err)
	err = e.StartJob(ctx, jid)
	tt.AssertNil(err)

	err = e.ExecuteJob(ctx)
	tt.AssertNil(err)
	job, err := e.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	job.Target.Components = job.Target.Components[1:]
	err = e.heavyUpdateParent(ctx, ctx, &job)
	tt.AssertNil(err)

	{
		// try error for test no block by
		ctx0, cancel := trait.WithCancelCauesContext(ctx)
		err1 := &trait.Error{
			Internal: trait.ECExit,
			Err:      context.Canceled,
			Detail:   "test cancel engine",
		}
		cancel(err1)
		err = e.heavyUpdateParent(ctx0, ctx0, &job)
		tt.Assert(err1, err)

		db := s.Store.(*mock.DbStoreFaker)

		err0 := &trait.Error{
			Err:      fmt.Errorf("GetComponentIns mock error"),
			Internal: trait.ECNULL,
		}
		db.ErrMap["GetComponentIns"] = err0
		err = e.heavyUpdateParent(ctx, ctx, &job)
		tt.Assert(err0, err)
	}
}
