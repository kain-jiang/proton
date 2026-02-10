package graph

import (
	"fmt"
	"slices"
	"sync"

	"taskrunner/pkg/component"
	"taskrunner/pkg/graph/task"
	"taskrunner/trait"

	"golang.org/x/mod/semver"
)

// ComponentNode a component  node in Plan
type ComponentNode struct {
	trait.Task
	Locker   sync.Locker
	chileNum int
	Children []*ComponentNode
	supper   []*ComponentNode
}

type compoentNodeList struct {
	next *compoentNodeList
	node *ComponentNode
}

// Plan application graph walker
type Plan struct {
	// nodes          map[string]*compoentNode
	Length         int
	layer          *compoentNodeList
	comsumerOffset int
	locker         sync.Locker
	con            *sync.Cond
}

// GetLoopEdge get edge between nodes
func GetLoopEdge(edges []trait.Edge, nodes []*trait.ComponentNode) []trait.Edge {
	index := make(map[string]*ComponentNode, len(nodes))
	res := make([]trait.Edge, 0)
	for _, t := range nodes {
		i := t
		index[i.Name] = &ComponentNode{
			Locker: &sync.Mutex{},
			Task: &task.Base{
				ComponentInsData: &trait.ComponentInstance{
					ComponentInstanceMeta: trait.ComponentInstanceMeta{
						Component: *t,
					},
				},
			},
		}
	}

	for _, e := range edges {
		from := e.From.Name
		to := e.To.Name
		f, ok0 := index[from]
		t, ok1 := index[to]
		if ok0 && ok1 {
			// res = append(res, e)s
			f.chileNum++
			f.Children = append(f.Children, t)
			t.supper = append(t.supper, f)
		}
	}

	sub := 1
	for sub != 0 {
		sub = 0
		for name, c := range index {
			if len(c.supper) == 0 {
				sub++
				delete(index, name)
				for _, child := range c.Children {
					child.supper = slices.DeleteFunc(child.supper, func(cn *ComponentNode) bool {
						return c.Component().Name == cn.Component().Name
					})
				}

			}
			if c.chileNum == 0 {
				sub++
				delete(index, name)
				for _, parent := range c.supper {
					parent.Children = slices.DeleteFunc(parent.Children, func(cn *ComponentNode) bool {
						return c.Component().Name == cn.Component().Name
					})
				}

			}
		}
	}

	for _, c := range index {
		for _, child := range c.Children {
			res = append(res, trait.Edge{
				From: *c.Component(),
				To:   *child.Component(),
			})
		}
	}

	return res
}

// ValidateTortuous the Tortuous nodes
func ValidateTortuous(edgs []trait.Edge, nodes []trait.Task) ([]*trait.ComponentNode, *trait.Error) {
	p, cs, err := NewFromGraph(edgs, nodes)
	if err != nil {
		return nil, err
	}
	index := make(map[string]trait.Task, len(nodes)+len(cs))
	for _, t := range cs {
		i := t.ComponentIns().Component
		index[i.Name] = t
	}
	for _, t := range nodes {
		i := t.ComponentIns().Component
		index[i.Name] = t
	}

	for {
		n := p.Next()
		if n == nil {
			break
		}
		p.Done(n)
		c := n.Component()
		delete(index, c.Name)
	}

	tortuous := make([]*trait.ComponentNode, 0, len(index))
	for _, c := range index {
		tortuous = append(tortuous, c.Component())
	}
	return tortuous, nil
}

// NewFromGraph reutrn graph's plan
func NewFromGraph(edgs []trait.Edge, tasks []trait.Task) (p *Plan, external []*task.Base, err *trait.Error) {
	p = &Plan{
		locker: &sync.Mutex{},
	}
	p.con = sync.NewCond(p.locker)
	nodes := map[string]*ComponentNode{}
	comIndex := map[string]trait.Task{}
	for _, t := range tasks {
		c := t.Component()
		if _, ok := comIndex[c.Name]; ok {
			err = &trait.Error{
				Err:      fmt.Errorf("不允许出现重复组件对象: %s", c.Name),
				Internal: trait.ErrComponentDup,
			}
			return
		}
		comIndex[c.Name] = t
	}

	indexOrNew := func(c trait.ComponentNode) (*ComponentNode, bool) {
		isExternal := false
		index := c.Name
		if n, ok := nodes[index]; ok {
			// change max version component
			if semver.Compare("v"+c.Version, "v"+n.Component().Version) == 1 {
				if t, ok := n.Task.(*task.Base); ok {
					t.ComponentInsData.Component.Version = c.Version
				}
			}
			return n, isExternal
		}

		realComponent, ok := comIndex[index]
		if realComponent == nil || !ok {
			c.ComponentDefineType = component.ComponentBaseType
			realComponent = &task.Base{
				ComponentInsData: &trait.ComponentInstance{
					ComponentInstanceMeta: trait.ComponentInstanceMeta{
						Component: c,
					},
				},
			}

			// external = append(external, c)
			isExternal = true
		}

		n := &ComponentNode{
			Locker:   &sync.Mutex{},
			Task:     realComponent,
			chileNum: 0,
			Children: make([]*ComponentNode, 0),
			supper:   make([]*ComponentNode, 0),
		}
		p.Length++
		nodes[index] = n
		return n, isExternal
	}

	for _, edge := range edgs {
		from, isExternal := indexOrNew(edge.From)
		if isExternal {
			external = append(external, from.Task.(*task.Base))
		}
		to, isExternal := indexOrNew(edge.To)
		if isExternal {
			external = append(external, to.Task.(*task.Base))
		}

		from.Children = append(from.Children, to)
		from.chileNum++
		to.supper = append(to.supper, from)
	}

	// orphan
	for _, t := range tasks {
		c := t.Component()
		index := c.Name
		if _, ok := nodes[index]; !ok {
			nodes[index] = &ComponentNode{
				Locker:   &sync.Mutex{},
				Task:     t,
				chileNum: 0,
			}
			p.Length++
		}
	}

	// deps plan
	for _, n := range nodes {
		if n.chileNum == 0 {
			p.layer = &compoentNodeList{
				node: n,
				next: p.layer,
			}
		}
	}
	return
}

// Done set the node done
func (p *Plan) Done(n *ComponentNode) {
	for _, s := range n.supper {
		s.Locker.Lock()
		s.chileNum--
		if s.chileNum == 0 {
			p.con.L.Lock()
			p.layer = &compoentNodeList{
				node: s,
				next: p.layer,
			}
			p.con.Signal()
			p.con.L.Unlock()
		}
		s.Locker.Unlock()
	}
}

// Next try get next node, if empty return nil
func (p *Plan) Next() *ComponentNode {
	p.locker.Lock()
	defer p.locker.Unlock()
	if p.layer == nil {
		return nil
	}

	n := p.layer
	p.layer = n.next
	p.comsumerOffset++
	return n.node
}

// NextBlock wait util next node or plan close
func (p *Plan) NextBlock() *ComponentNode {
	p.con.L.Lock()
	defer p.con.L.Unlock()
	length := p.Length
	for p.layer == nil && p.comsumerOffset < length {
		p.con.Wait()
	}

	if p.comsumerOffset >= length {
		p.con.Broadcast()
		return nil
	}

	p.comsumerOffset++
	n := p.layer
	p.layer = n.next
	if p.comsumerOffset >= length {
		p.con.Broadcast()
	}

	return n.node
}

// Close close plan , wake up NextBlock routine
func (p *Plan) Close() {
	p.con.L.Lock()
	defer p.con.L.Unlock()
	p.comsumerOffset = p.Length
	p.con.Broadcast()
}
