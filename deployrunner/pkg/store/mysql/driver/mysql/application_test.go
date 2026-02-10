package store_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"taskrunner/pkg/app/builder"
	"taskrunner/trait"

	testdata "taskrunner/test"
	testcharts "taskrunner/test/charts"
)

func getTestAppliation(t *testing.T) trait.Application {
	// apf := testdata.TestAPP
	// testWantError(t, nil, err)
	// defer apf.Close()
	apf := bytes.NewReader(testdata.TestAPP)
	buf := bytes.NewBuffer(nil)
	tt := testdata.TestingT{T: t}
	cfg, err := builder.LoadConfiguration(apf)
	tt.AssertNil(err)

	repo := &testcharts.MemoryHelmRepoMock{
		RepoName: "test",
		FS:       testdata.TestCharts,
	}

	b, err := builder.NewApplicationBuilder(&cfg, buf, io.Discard, repo)
	tt.AssertNil(err)

	err = b.Build(context.Background())

	tt.AssertNil(err)
	bs := buf.Bytes()
	a, _, err := builder.ParseApplication(bytes.NewReader(bs))
	tt.AssertNil(err)
	a.Component[0].Type = "test"
	return a
}

func TestApplication(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := testdata.TestingT{T: t}
	a := getTestAppliation(t)
	ctx := context.Background()
	aid, err := s.InsertAPP(ctx, a)
	tt.AssertNil(err)

	_, err = s.InsertAPP(ctx, a)
	tt.AssertError(trait.ErrUniqueKey, err)

	a0, err := s.GetAPP(ctx, aid)
	tt.AssertNil(err)
	tt.Assert(a0.AName, a.AName)
	tt.Assert(a0.Version, a.Version)
	tt.Assert(10, len(a0.Component))
	// tt.Assert(a0.ConfigSchema, json.RawMessage(nil))

	as, err := s.SearchAPP(ctx, 10, -1, a.AName)
	tt.AssertNil(err)
	tt.Assert(1, len(as))

	as, err = s.SearchAPP(ctx, 10, as[0].AID, a.AName)
	tt.AssertNil(err)
	tt.Assert(0, len(as))

	as, err = s.ListAPP(ctx, 10, -1)
	tt.AssertNil(err)
	tt.Assert(1, len(as))

	as, err = s.ListAPP(ctx, 10, as[0].AID)
	tt.AssertNil(err)
	tt.Assert(0, len(as))

	for _, c := range a0.Component {
		meta, err := s.GetAPPComponent(ctx, c.CID)
		tt.AssertNil(err)
		tt.Assert(meta.Name, c.Name)
		tt.Assert(meta.ComponentNode, c.ComponentNode)
		tt.Assert(meta.Spec, c.Spec)
		tt.Assert(meta.Type, c.Type)
	}

	buf := bytes.NewBufferString("test log version")
	for i := 0; i < 1024; i++ {
		buf.WriteString("test")
	}

	a.Version = buf.String()
	_, err = s.InsertAPP(ctx, a)
	tt.AssertError(trait.ErrParam, err)

	{
		a.Version = "test"
		a.Dependence = []trait.AppDepMeta{
			{
				AName:   "dep0",
				Version: "1.0.0",
			},
			{
				AName:   "dep1",
				Version: "1.0.1",
			},
		}
		aid0, err := s.InsertAPP(ctx, a)
		tt.AssertNil(err)
		a1, err := s.GetAPP(ctx, aid0)
		tt.AssertNil(err)
		tt.Assert(len(a.Dependence), len(a1.Dependence))

		tt.AssertNil(s.UpdateAppDependence(ctx, a))
	}

	{

		a.Version = "testwork"
		aid, err = s.InsertAPP(ctx, a)
		a.AID = aid
		tt.AssertNil(err)
		ains := &trait.ApplicationInstance{
			Application: a,
		}
		err = s.WorkAppIns(ctx, ains)
		tt.AssertNil(err)
		as, err = s.ListSystemAPPNoWorked(ctx, 10, -1, ains.SID)
		tt.AssertNil(err)
		tt.Assert(0, len(as))

		a.AName = "nowork"
		aid, err = s.InsertAPP(ctx, a)
		tt.AssertNil(err)
		as, err = s.ListSystemAPPNoWorked(ctx, 10, -1, ains.SID)
		tt.AssertNil(err)
		tt.Assert(aid, as[0].AID)
	}

	{
		err = s.DeleteAPP(ctx, aid)
		tt.AssertNil(err)
		_, err = s.GetAPP(ctx, -1)
		tt.AssertError(trait.ErrNotFound, err)
	}
}
