package store_test

import (
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

func TestComponentInstance(t *testing.T) {
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
	tt.AssertNil(err)
	_, err = s.InsertAPPIns(ctx, ains)
	tt.AssertNil(err)

	cid := ains.Components[0].CID
	cins, err := s.GetComponentIns(ctx, cid)
	tt.Assert("test", cins.Component.Type)
	tt.AssertNil(err)
	oldrevision := cins.Revission
	err = s.UpdateComponentInsStatus(ctx, cid, trait.AppDoingStatus, cins.Revission, -1, -1)
	tt.AssertNil(err)
	err = s.UpdateComponentInsStatus(ctx, cid, trait.AppDoingStatus, cins.Revission, 1, -1)
	tt.AssertError(trait.ErrComponentInstanceRevission, err)
	cins, err = s.GetComponentIns(ctx, cid)
	tt.AssertNil(err)
	tt.Assert(cins.Status, trait.AppDoingStatus)
	tt.Assert(oldrevision+1, cins.Revission)

	err = s.WorkComponentIns(ctx, cins)
	tt.AssertNil(err)

	err = s.WorkComponentIns(ctx, cins)
	tt.AssertNil(err)

	cins, err = s.GetWorkComponentIns(ctx, sid, cins.Component)
	tt.AssertNil(err)
	cs, err := s.ListWorkComponentIns(ctx, trait.WorkCompFilter{
		Aname: cins.APPName,
		Sid:   cins.System.SID,
	})
	tt.AssertNil(err)
	tt.Assert(1, len(cs))
	tt.Assert(trait.AppDoingStatus, cins.Status)
	err = s.LayoffComponentIns(ctx, cins.CID)
	tt.AssertNil(err)

	_, err = s.GetWorkComponentIns(ctx, sid, cins.Component)
	tt.AssertError(trait.ErrNotFound, err)

	{
		err = s.LockComponent(ctx, sid, 1, cins.Component)
		tt.AssertNil(err)
		ctx0, cancel := trait.WithTimeoutCauseContext(ctx, 2*time.Second, &trait.Error{
			Internal: trait.ECTimeout,
			Err:      fmt.Errorf("timeout for lock component"),
		})
		defer cancel()
		err = s.LockComponent(ctx, sid, 1, cins.Component)
		tt.AssertNil(err)

		err = s.LockComponent(ctx0, sid, 1, cins.Component)
		tt.AssertNil(err)

		err = s.LockComponent(ctx0, sid, 2, cins.Component)
		tt.AssertError(trait.ECTimeout, err)

		err = s.UnlockComponent(ctx, sid, 1, cins.Component)
		tt.AssertNil(err)

		err = s.LockComponent(ctx, sid, 2, cins.Component)
		tt.AssertNil(err)

		err = s.UnlockJobComponent(ctx, 2)
		tt.AssertNil(err)
	}
}
