package task

import (
	"fmt"

	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component"
	"taskrunner/trait"
)

// TODO a factory registry component validator and component task

// NewTask create a component instace task
func NewTask(meta *trait.ComponentMeta, systemContext *cluster.SystemContext) (trait.Task, *trait.Error) {
	ins := &trait.ComponentInstance{
		ComponentInstanceMeta: trait.ComponentInstanceMeta{
			Component: meta.ComponentNode,
		},
	}
	switch meta.ComponentDefineType {
	case HelmServiceType, HelmTaskType:
		hc := &component.HelmComponent{
			ComponentMeta: *meta,
		}
		if err := hc.Decode(meta.Spec); err != nil {
			err := fmt.Errorf("the component instance %s:%s decode error %s", meta.Name, meta.Version, err.Error())
			return nil, &trait.Error{
				Internal: trait.ECComponentDefined,
				Err:      err,
				Detail:   fmt.Sprintf("decode special defined: %s", string(meta.Spec)),
			}
		}
		return &HelmTask{
			System:        systemContext,
			HelmComponent: hc,
			Base: Base{
				ComponentInsData: ins,
			},
		}, nil
	case holeTaskType, component.ComponentHelmAddtionalType:
		c := &HoleTask{
			System: systemContext,
			Base: Base{
				ComponentInsData: ins,
			},
		}
		return c, nil
	case baseTaskType:
		return &Base{
			ComponentInsData: ins,
		}, nil
	case protonResourceType:
		return newProtonResourceTask(ins, systemContext), nil
	default:
		err := fmt.Errorf("component %s with component type  %s not support", meta.Name, meta.ComponentDefineType)
		return nil, &trait.Error{
			Err:      err,
			Internal: trait.ErrComponentTypeNotDefined,
			Detail:   "create task for component",
		}
	}
}

// NewTasks generate task from application without instance config
// warn shuold set task config before use it to install
func NewTasks(app *trait.Application, systemContext *cluster.SystemContext) ([]trait.Task, *trait.Error) {
	ts := make([]trait.Task, 0, len(app.Component))
	for _, c := range app.Component {
		// repo attribute reserve for futrue
		t, err := NewTask(c, systemContext)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}
