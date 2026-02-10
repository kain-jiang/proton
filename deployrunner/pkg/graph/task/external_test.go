package task

import (
	"context"
	"testing"

	"taskrunner/test"
	"taskrunner/trait"
)

func TestBase(t *testing.T) {
	b := Base{
		ComponentInsData: &trait.ComponentInstance{
			ComponentInstanceMeta: trait.ComponentInstanceMeta{
				Component: trait.ComponentNode{},
			},
			Attribute: make(map[string]interface{}),
			Config:    map[string]interface{}{},
		},
	}
	tt := test.TestingT{T: t}
	b.SetTopology(nil)
	tt.AssertError(trait.ECBaseNode, b.Install(context.Background()))
	tt.AssertError(trait.ECBaseNode, b.Uninstall(context.Background()))
	if b.ComponentIns() != b.ComponentInsData {
		t.Fatal()
	}
	if b.Component() != &b.ComponentInsData.Component {
		t.Fatal()
	}
	attr := b.Attribute()
	if attr == nil {
		t.Fatal(&attr)
	}
	cins := &trait.ComponentInstance{}
	err := b.SetComponentIns(cins)
	tt.AssertNil(err)
	tt.Assert(b.ComponentInsData, cins)
}
