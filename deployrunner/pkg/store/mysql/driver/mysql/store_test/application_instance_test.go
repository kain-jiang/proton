package store_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"taskrunner/pkg/app"
	"taskrunner/test"
	"taskrunner/trait"
)

func TestAPPlicationIns(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	randNum, err0 := rand.Int(rand.Reader, big.NewInt(math.MaxInt))
	tt.Assert(nil, err0)
	ss := trait.System{
		NameSpace: fmt.Sprintf("test_%s_%d", test.GetSourceBranch(), int(randNum.Int64())),
		SName:     "test",
	}
	ctx := context.Background()
	sid, err := s.InsertSystemInfo(ctx, ss)
	tt.AssertNil(err)
	ss.SID = sid

	a := getTestAppliation(t)
	aid, err := s.InsertAPP(ctx, a)
	tt.AssertNil(err)
	a.AID = aid

	ains, err := app.NewAPPIns(nil, ss, a)
	ains.Components[0].Timeout = 10
	tt.AssertNil(err)
	id, err := s.InsertAPPIns(ctx, ains)
	ains.ID = id
	tt.AssertNil(err)

	ains0, err := s.GetAPPIns(ctx, id)
	tt.AssertNil(err)
	tt.Assert(ains0.AName, ains.AName)
	tt.Assert(ains0.Version, ains.Version)
	tt.Assert(len(ains.Components), len(ains0.Components))
	tt.Assert(10, ains0.Components[0].Timeout)
	tt.Assert(ains.Components[0].Acid, ains0.Components[0].Acid)

	ains.Components[0].Status = trait.AppDoingStatus
	ains.AppConfig = map[string]interface{}{
		"a": 1,
	}
	comment := "test"
	ains.Comment = comment
	err = s.UpdateAPPInsConfig(ctx, *ains)
	tt.AssertNil(err)
	tt.Assert(ains.Comment, comment)
	ains, err = s.GetAPPIns(ctx, ains.ID)
	tt.AssertNil(err)
	tt.Assert(1.0, ains.AppConfig["a"])
	ains.Status = trait.AppDoingStatus
	err = s.UpdateAPPInsStatus(ctx, ains.ID, ains.Status, 1, int(time.Now().Unix()), -1)
	tt.AssertNil(err)
	tt.AssertNil(s.UpdateAPPInsOperateType(ctx, ains.ID, trait.JobDeleteOType))

	ains0, err = s.GetAPPIns(ctx, id)
	tt.AssertNil(err)
	tt.Assert(ains0.Status, ains.Status)
	tt.Assert(ains0.Components[0].Status, ains.Components[0].Status)
	tt.Assert(1, ains0.Onwer)
	tt.Assert(trait.JobDeleteOType, ains0.OType)

	err = s.WorkAppIns(ctx, ains)
	tt.AssertNil(err)
	err = s.WorkAppIns(ctx, ains)
	tt.AssertNil(err)
	ains0, err = s.GetWorkAPPIns(ctx, ains.AName, sid)
	tt.AssertNil(err)
	tt.Assert(ains0.ID, ains.ID)
	tt.Assert(ains.AID, ains0.AID)
	// t.Log(ains0.AID)

	{
		as, err := s.ListWorkAPPIns(ctx, &trait.AppInsFilter{
			Offset: 0,
			Sid:    -1,
			Limit:  10,
		})
		tt.AssertNil(err)
		tt.Assert(1, len(as))
	}

	as, err := s.ListWorkAPPIns(ctx, &trait.AppInsFilter{
		Offset: 0,
		Sid:    sid,
		Limit:  10,
	})
	tt.AssertNil(err)
	tt.Assert(1, len(as))

	tt.AssertNil(err)
	tt.Assert(1, len(as))
	c, err := s.CountWorkAppIns(ctx, &trait.AppInsFilter{Offset: 0, Sid: sid, Limit: 10})
	tt.AssertNil(err)
	tt.Assert(1, c)
	as, err = s.ListWorkAPPIns(ctx, &trait.AppInsFilter{
		Offset: 1,
		Sid:    sid,
		Limit:  10,
	})
	tt.AssertNil(err)
	tt.Assert(0, len(as))

	// work the same application again, should idempotent
	err = s.WorkAppIns(ctx, ains)
	tt.AssertNil(err)

	err = s.LayOffAPPIns(ctx, ains)
	tt.AssertNil(err)
	_, err = s.GetWorkAPPIns(ctx, ains.AName, sid)
	tt.AssertError(trait.ErrNotFound, err)
	as, err = s.ListWorkAPPIns(ctx, &trait.AppInsFilter{
		Offset: 0,
		Sid:    sid,
		Limit:  10,
	})
	tt.AssertNil(err)
	tt.Assert(0, len(as))

	buf := bytes.NewBufferString("test log version")
	for i := 0; i < 1024; i++ {
		buf.WriteString("test")
	}
	ains.AName = "qweqweqweqwe"
	ains.Version = buf.String()
	_, err = s.InsertAPPIns(ctx, ains)
	tt.AssertError(trait.ErrParam, err)
}

func TestAppLocker(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	aname := "test"
	tt.AssertNil(s.LockApp(ctx, 1, 1, aname))
	tt.AssertNil(s.LockApp(ctx, 1, 1, aname))
	tt.AssertNil(s.LockApp(ctx, 2, 1, aname))
	tt.AssertNil(s.LockApp(ctx, 2, 1, aname))
	tt.AssertNil(s.UnlockApp(ctx, 1, 1, aname))
	tt.AssertNil(s.UnlockApp(ctx, 1, 1, aname))
	tt.AssertNil(s.LockApp(ctx, 1, 1, aname))
	ctx0, cancel := trait.WithTimeoutCauseContext(ctx, 500*time.Millisecond, &trait.Error{
		Internal: trait.ECTimeout,
	})
	defer cancel()
	err := s.LockApp(ctx0, 2, 2, aname)
	tt.AssertError(trait.ECTimeout, err)
}
