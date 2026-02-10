package app

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"taskrunner/pkg/helm"
	"taskrunner/test"
	"taskrunner/test/mock"
	"taskrunner/trait"

	"github.com/sirupsen/logrus"
)

var getTestAppliationBytes = mock.GetTestAppliationBytes

func testStoreInstance(_ *testing.T) Store {
	repo := mock.HelmRepoMock{
		RepoName: "test",
		Chart:    map[string][]byte{},
		WantErr:  nil,
	}
	s := NewStore(logrus.New(), &mock.DbStoreFaker{
		ErrMap:       make(map[string]*trait.Error),
		AppLangCache: make(map[string]string),
	}, helm.NewHelmIndexRepo(&repo))
	s.Log.SetReportCaller(true)
	return *s
}

func TestUploadApplicationPackage(t *testing.T) {
	bs := getTestAppliationBytes(t)
	s := testStoreInstance(t)
	ctx := context.Background()
	tt := test.TestingT{T: t}

	_, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)
	repo := mock.HelmRepoMock{
		RepoName: "test",
		Chart:    make(map[string][]byte),
	}
	s.HelmRepo = helm.NewHelmIndexRepo(&repo)
	err = &trait.Error{
		Internal: trait.ErrHelmRepoNoFound,
	}
	repo.WantErr = err
	_, err = s.UploadApplicationPackage(ctx, bytes.NewReader([]byte{'q'}))
	if !trait.IsInternalError(err, trait.ErrApplicationFile) {
		t.Fatal(err)
	}

	_, err = s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertError(trait.ErrHelmRepoNoFound, err)

	err0 := &trait.Error{
		Internal: trait.ECNULL,
		Err:      fmt.Errorf("test mock"),
	}
	repo.WantErr = err0
	_, err = s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.Assert(err0, err)

	db := s.Store.(*mock.DbStoreFaker)
	err0 = &trait.Error{
		Internal: trait.ECNULL,
		Err:      fmt.Errorf("test mock"),
	}
	db.WantErr = err0
	repo.WantErr = nil
	_, err = s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.Assert(err0, err)
}

func TestNewJobRecord(t *testing.T) {
	bs := getTestAppliationBytes(t)
	s := testStoreInstance(t)
	ctx := context.Background()
	tt := test.TestingT{T: t}

	aid, err := s.UploadApplicationPackage(ctx, bytes.NewReader(bs))
	tt.AssertNil(err)
	db := s.Store.(*mock.DbStoreFaker)
	sid, err := s.InsertSystemInfo(ctx, trait.System{
		NameSpace: "test",
	})
	tt.AssertNil(err)
	jid, err := s.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)

	job, err := s.GetJobRecord(ctx, jid)
	tt.AssertNil(err)

	_, err = s.GetJobRecord(ctx, -1)
	tt.AssertError(trait.ErrNotFound, err)

	tt.Assert(10, len(job.Target.Components))
	tt.Assert(aid, job.Target.AID)

	db.SystemCache[sid].WorkAPP = map[string]*trait.ApplicationInstance{
		job.Target.Application.AName: job.Target,
	}

	newjob := job.Target
	newjob.AppConfig = map[string]interface{}{
		"a": 1,
	}
	err = s.SetJobConfig(ctx, job.Target.ID, newjob)
	tt.AssertNil(err)
	if newjob != job.Target {
		t.FailNow()
	}

	job, err = s.GetJobRecord(ctx, jid)
	tt.AssertNil(err)
	tt.Assert(job.Target.AppConfig["a"], 1)

	job.Target.Status = trait.AppWaitingStatus
	newjob = &trait.ApplicationInstance{}
	err = s.SetJobConfig(ctx, jid, job.Target)
	tt.AssertError(trait.ErrJobExecuting, err)
	if newjob == job.Target {
		t.FailNow()
	}

	job.Target.Components[0].Timeout = 1
	_, err = s.NewJobRecord(ctx, aid, sid)
	tt.AssertNil(err)
}

func TestAppLang(t *testing.T) {
	ss := testStoreInstance(t)
	ctx := context.Background()
	tt := test.TestingT{T: t}
	s := ss

	lang := "zh-cn"
	aname := "test"
	alias, err := s.GetAppLang(ctx, lang, aname, "")
	tt.AssertNil(err)
	tt.Assert(alias, aname)

	aname0 := ss.GetAname(lang, alias, "")
	tt.Assert(aname, aname0)

	tt.AssertNil(s.InsertAppLang(ctx, lang, aname, alias, ""))
	alias = "测试"
	tt.AssertNil(s.InsertAppLang(ctx, lang, aname, alias, ""))
	a, err := s.GetAppLang(ctx, lang, aname, "")
	tt.AssertNil(err)
	tt.Assert(a, alias)
}
