package store_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"taskrunner/test"
	"taskrunner/trait"
)

func TestEdgeCrossTransaction(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	ctx := context.Background()
	tt := test.TestingT{T: t}

	testcase := [][2]int{
		{1, 2},
		{1, 3},
		{0, 2},
		{0, 3},
		{5, 6},
	}
	err := s.AddEdge(ctx, testcase[0][0], testcase[0][1])
	tt.AssertNil(err)

	for _, e := range testcase {
		err = s.AddEdge(ctx, e[0], e[1])
		tt.AssertNil(err)
	}
	tx, err := s.Begin(ctx)
	tt.AssertNil(err)
	err = tx.DeleteEdgeFrom(ctx, 5)
	tt.AssertNil(err)
	err = tx.AddEdge(ctx, 7, 8)
	tt.AssertNil(err)
	err = tx.AddEdge(ctx, 7, 8)
	tt.AssertNil(err)
	t.Log("tx")

	ctx0, cancel := trait.WithTimeoutCauseContext(ctx, 15*time.Second, nil)
	defer cancel()
	tx0, err := s.Begin(ctx0)
	tt.AssertNil(err)
	t.Log("tx0")
	go func() {
		start := time.Now()
		time.Sleep(5 * time.Second)
		tt.AssertNil(tx.Commit())
		cost := time.Since(start).Seconds()
		t.Log("tx end, cost: ", cost)
	}()
	start := time.Now()

	err = tx0.ChangeEdgeto(ctx, 6, 0)
	tt.AssertNil(err)
	t.Log("commit")

	tt.AssertNil(tx0.Commit())
	cost := time.Since(start).Seconds()
	t.Log("tx0 end, cost: ", cost)
}

func TestEdgeAddOuterChildEdge(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	ctx := context.Background()
	tt := test.TestingT{T: t}
	cins := &trait.ComponentInstance{
		ComponentInstanceMeta: trait.ComponentInstanceMeta{
			System: trait.System{
				SID: 1,
			},
			Component: trait.ComponentNode{
				Name: "test",
			},
			CID: 2,
		},
	}
	{
		err := s.WorkComponentIns(ctx, cins)
		tt.AssertNil(err)
		err = s.AddOuterChildEdge(ctx, 1, cins.System.SID, cins.Component)
		tt.AssertNil(err)
		c, err := s.CountEdgeTo(ctx, cins.CID)
		tt.AssertNil(err)
		test.Assert(t, 1, c)
		err = s.LayoffComponentIns(ctx, cins.CID)
		tt.AssertNil(err)
		err = s.DeleteEdge(ctx, 1, cins.CID)
		tt.AssertNil(err)
		err = s.AddOuterChildEdge(ctx, 1, cins.System.SID, cins.Component)
		tt.AssertNil(err)
		c, err = s.CountEdgeTo(ctx, cins.CID)
		tt.AssertNil(err)
		test.Assert(t, 0, c)
	}
}

func TestEdge(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	ctx := context.Background()
	tt := test.TestingT{T: t}

	testcase := [][2]int{
		{1, 2},
		{1, 3},
		{0, 2},
		{2, 3},
		{5, 6},
	}
	err := s.AddEdge(ctx, testcase[0][0], testcase[0][1])
	tt.AssertNil(err)

	for _, e := range testcase {
		err = s.AddEdge(ctx, e[0], e[1])
		tt.AssertNil(err)
	}

	{
		// test GetPointTo
		froms, err := s.GetPointTo(ctx, 2)
		tt.AssertNil(err)
		sort.Slice(froms, func(i, j int) bool {
			return froms[i] < froms[j]
		})
		get := make([]interface{}, 0, len(froms))
		for _, i := range froms {
			get = append(get, i)
		}
		tt.AssertArray([]interface{}{0, 1}, get)
	}

	{
		// test GetPointFrom
		tos, err := s.GetPointFrom(ctx, 0)
		tt.AssertNil(err)
		sort.Slice(tos, func(i, j int) bool {
			return tos[i] < tos[j]
		})
		get := make([]interface{}, 0, len(tos))
		for _, i := range tos {
			get = append(get, i)
		}
		tt.AssertArray([]interface{}{2}, get)

	}

	count, err := s.CountEdgeTo(ctx, 6)
	tt.AssertNil(err)
	tt.Assert(1, count)

	// delete 5->6
	err = s.DeleteEdge(ctx, testcase[4][0], testcase[4][1])
	tt.AssertNil(err)
	err = s.DeleteEdge(ctx, testcase[4][0], testcase[4][1])
	tt.AssertNil(err)
	count, err = s.CountEdgeTo(ctx, 6)
	tt.AssertNil(err)
	tt.Assert(0, count)

	err = s.ChangeEdgeto(ctx, 3, 5)
	tt.AssertNil(err)

	// ids, err := s.GetChangeEdgeToConflictStmt(ctx, 5, 2)
	// tt.AssertNil(err)
	// tt.Assert(1, len(ids))
	err = s.ChangeEdgeto(ctx, 5, 2)
	tt.AssertNil(err)

	count, err = s.CountEdgeTo(ctx, 5)
	tt.AssertNil(err)
	tt.Assert(0, count)

	err = s.DeleteEdgeFrom(ctx, 5)
	tt.AssertNil(err)
	count, err = s.CountEdgeTo(ctx, 6)
	tt.AssertNil(err)
	tt.Assert(0, count)
}
