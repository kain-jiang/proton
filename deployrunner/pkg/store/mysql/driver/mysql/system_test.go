package store_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"testing"

	"taskrunner/test"
	"taskrunner/trait"
)

func TestSystem(t *testing.T) {
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
	_, err := s.ListSystemInfo(ctx, 1, 0)
	tt.AssertNil(err)
	sid, err := s.InsertSystemInfo(ctx, ss)
	tt.AssertNil(err)

	s0, err := s.GetSystemInfo(ctx, sid)
	tt.AssertNil(err)
	tt.Assert(s0.NameSpace, ss.NameSpace)
	tt.Assert(s0.SName, ss.SName)
	tt.AssertNil(s0.Config)

	// s0.NameSpace = "testnamechange"
	s0.SName = "testchangeSName"
	s0.Config = map[string]interface{}{
		"test": 1.0,
	}
	err = s.UpdateSystemInfo(ctx, *s0)
	tt.AssertNil(err)

	err = s.UpdateSystemInfo(ctx, *s0)
	tt.AssertNil(err)

	s1, err := s.GetSystemInfo(ctx, sid)
	tt.AssertNil(err)
	tt.Assert(s0.SName, s1.SName)
	tt.Assert(s0.Config, s1.Config)
	// tt.Assert(s0.NameSpace, s1.NameSpace)

	_, err = s.GetSystemInfo(ctx, -1)
	tt.AssertError(trait.ErrNotFound, err)

	buf := bytes.NewBufferString("test long namespace")
	for i := 0; i < 1024; i++ {
		buf.WriteString("test")
	}
	ss.NameSpace = buf.String()
	_, err = s.InsertSystemInfo(ctx, ss)
	tt.AssertError(trait.ErrParam, err)

	ss.NameSpace = "testchangeSName"
	_, err = s.InsertSystemInfo(ctx, ss)
	tt.AssertNil(err)
	count, err := s.CountSystemInfo(ctx)
	tt.AssertNil(err)
	tt.Assert(2, count)

	ss.NameSpace = "testchangeSName"
	_, err = s.InsertSystemInfo(ctx, ss)
	tt.AssertError(trait.ErrUniqueKey, err)

	ss.NameSpace = "qweasdqliuyew"
	_, err = s.InsertSystemInfo(ctx, ss)
	tt.AssertError(trait.ErrUniqueKey, err)

	slist, err := s.ListSystemInfo(ctx, 3, 0)
	tt.AssertNil(err)
	tt.Assert(2, len(slist))

	ss0, err := s.GetSystemInfoByName(ctx, ss.SName)
	tt.AssertNil(err)
	tt.Assert(ss0.SName, ss.SName)

	_, err = s.GetSystemInfoByName(ctx, "ajsydoiuqasd")
	tt.AssertError(trait.ErrNotFound, err)
}
