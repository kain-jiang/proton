package graph

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"taskrunner/pkg/graph/task"
	"taskrunner/test"
	"taskrunner/trait"
)

func testGetEdges(es [][2]int, componentLength int) ([]trait.Edge, []trait.Task) {
	length := componentLength
	cs := make([]trait.Task, 0, length)
	for i := 0; i < length; i++ {
		cs = append(cs, &task.Base{
			ComponentInsData: &trait.ComponentInstance{
				ComponentInstanceMeta: trait.ComponentInstanceMeta{
					Component: trait.ComponentNode{
						Name:    fmt.Sprintf("c%d", i),
						Version: "0.0.0",
					},
				},
			},
		})
	}
	edges := make([]trait.Edge, 0)
	for _, e := range es {
		edges = append(edges, trait.Edge{
			From: *cs[e[0]].Component(),
			To:   *cs[e[1]].Component(),
		})
	}
	return edges, cs
}

func TestNewGraph(t *testing.T) {
	tt := test.TestingT{T: t}
	edgesIndex := [][2]int{
		{0, 1},
		{1, 2},
		{1, 0},
		{1, 3},
		{2, 3},
	}
	edges, nodes := testGetEdges(edgesIndex, 5)
	_, _, err := NewFromGraph(edges, nodes)
	tt.AssertNil(err)
	nodes = append(nodes, &task.Base{
		ComponentInsData: &trait.ComponentInstance{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				Component: trait.ComponentNode{
					Name:    "c4",
					Version: "0.1.0",
				},
			},
		},
	})
	_, _, err = NewFromGraph(edges, nodes)
	tt.AssertError(trait.ErrComponentDup, err)
}

func TestPlanIterator(t *testing.T) {
	// 0->1->2->3
	// 0<-1->3
	tt := test.TestingT{T: t}
	edgesIndex := [][2]int{
		{0, 1},
		{1, 2},
		{1, 0},
		{1, 3},
		{2, 3},
		{4, 0},
	}
	edges, _ := testGetEdges(edgesIndex, 5)
	csLoop, _ := ValidateTortuous(edges, nil)
	if len(csLoop) != 3 {
		t.Fatal(csLoop)
	}

	edges = GetLoopEdge(edges, csLoop)
	if len(edges) != 2 {
		t.Fatal(edges)
	}

	// 0->1->2->3
	// 1->3->4
	edgesIndex = [][2]int{
		{0, 1},
		{1, 2},
		{1, 3},
		{2, 3},
		{3, 4},
	}
	edges, cs := testGetEdges(edgesIndex, 5)
	// 0->4 0.1.0
	edges = append(edges, trait.Edge{
		From: edges[0].From,
		To: trait.ComponentNode{
			Name:    "c4",
			Version: "0.1.0",
		},
	})
	// 0->4 0.2.0
	edges = append(edges, trait.Edge{
		From: edges[0].From,
		To: trait.ComponentNode{
			Name:    "c4",
			Version: "0.2.0",
		},
	})

	p, _, err := NewFromGraph(edges, cs)
	tt.AssertNil(err)
	// testWantError(t, nil, err)
	want := []int{4}
	for _, i := range want {
		c := p.Next()
		if c.Component().Name != cs[i].Component().Name || c.Component().Version != "0.2.0" {
			t.Fatal(c.Component(), i, cs[i].Component())
		}
		p.Done(c)
	}

	want = []int{3, 2, 1, 0}
	for _, i := range want {
		c := p.Next()
		if c.Component().Name != cs[i].Component().Name {
			t.Fatal(c.Component(), i, cs[i].Component())
		}
		p.Done(c)
	}
	if c := p.Next(); c != nil {
		t.Fatal(c)
	}
}

func TestPlanAsyncNoRing(t *testing.T) {
	// 0->1->2
	//    1->3->2
	edgesIndex := [][2]int{
		{0, 1},
		{1, 2},
		{1, 3},
		{3, 2},
	}
	edges, _ := testGetEdges(edgesIndex, 5)
	p, _, _ := NewFromGraph(edges, nil)
	// testWantError(t, nil, err)

	parrell := 10

	wg := &sync.WaitGroup{}
	wg.Add(parrell)
	want := map[string]int{
		"c0": 0,
		"c1": 0,
		"c2": 0,
		"c3": 0,
	}
	lock := &sync.Mutex{}
	for i := 0; i < parrell; i++ {
		go func() {
			defer wg.Done()
			for {
				c := p.NextBlock()
				if c == nil {
					return
				}
				p.Done(c)
				lock.Lock()
				want[c.Component().Name]++
				lock.Unlock()
			}
		}()
	}

	wg.Wait()
	for _, c := range want {
		if c != 1 {
			t.Fatal(want)
		}
	}
}

func TestPlanAsyncClose(t *testing.T) {
	edgesIndex := [][2]int{
		{0, 1},
		{1, 2},
		{1, 0},
		{1, 3},
		{2, 3},
	}
	edges, _ := testGetEdges(edgesIndex, 5)
	p, _, _ := NewFromGraph(edges, nil)
	// p, err := NewPlan(edges, cs)

	count := int64(0)
	parrell := 10

	wg := &sync.WaitGroup{}
	wg.Add(parrell)
	for i := 0; i < parrell; i++ {
		go func() {
			defer wg.Done()
			for {
				c := p.NextBlock()
				if c == nil {
					return
				}
				p.Done(c)
				atomic.AddInt64(&count, 1)
			}
		}()
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		p.Close()
	}()

	wg.Wait()
	if count != 2 {
		t.Fatal(count)
	}
}
