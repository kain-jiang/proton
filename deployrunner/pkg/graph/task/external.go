package task

import (
	"context"
	"fmt"

	"taskrunner/pkg/log"
	"taskrunner/trait"
)

// ErrBaseNode internal error for base task
var ErrBaseNode = fmt.Errorf("base node needn't deal")

// Base a external task node in application
// only use to cache component data
type Base struct {
	// appConfig           config
	ComponentInsData *trait.ComponentInstance
	// OldComponentInsData *trait.ComponentInstance
	Topology []*trait.ComponentInstance
	Log      *log.TaskLogger
}

// SetComponentIns impl task interface
func (h *Base) SetComponentIns(cins *trait.ComponentInstance) *trait.Error {
	com := h.ComponentInsData.ComponentInstanceMeta.Component
	cins.ComponentInstanceMeta.Component = com
	h.ComponentInsData = cins
	return nil
}

func (h *Base) WithLog(log *log.TaskLogger) {
	h.Log = log
}

// Install impl task interface
func (h *Base) Install(ctx context.Context) *trait.Error {
	// neen't done
	return &trait.Error{
		Internal: trait.ECBaseNode,
	}
}

// Uninstall impl task interface
func (h *Base) Uninstall(ctx context.Context) *trait.Error {
	// neen't done
	return &trait.Error{
		Internal: trait.ECBaseNode,
	}
}

// Component imply task interface
func (h *Base) Component() *trait.ComponentNode {
	return &h.ComponentInsData.Component
}

// ComponentIns imply task interface
func (h *Base) ComponentIns() *trait.ComponentInstance {
	return h.ComponentInsData
}

// Attribute no need
func (h *Base) Attribute() config {
	return h.ComponentInsData.Attribute
}

// SetTopology set topolofy
func (h *Base) SetTopology(cs []*trait.ComponentInstance) {
	h.Topology = cs
}

// // OldComponentIns impl task interface
// func (h *Base) OldComponentIns() *trait.ComponentInstance {
// 	return h.OldComponentInsData
// }
