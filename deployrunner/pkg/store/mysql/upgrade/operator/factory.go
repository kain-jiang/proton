package operator

import (
	"context"
	"fmt"

	"taskrunner/pkg/store/mysql/upgrade/store"
	"taskrunner/pkg/store/mysql/upgrade/trait"
	ctrait "taskrunner/trait"
)

// ObjectStore plan obj store, some operator may compose  other operator
type ObjectStore map[string]any

// Executor 算子执行器
type Executor interface {
	Execute(ctx context.Context, w *trait.WorkEnv, s store.Store, pm trait.PlanProcess) trait.Error
}

var executorBuilder map[string]func(objs ObjectStore, op trait.Operator) (Executor, trait.Error)

// NewExecutor 创建算子执行
func NewExecutor(objs ObjectStore, op trait.Operator) (Executor, trait.Error) {
	b := executorBuilder[op.Command]
	if b != nil {
		return b(objs, op)
	}
	return nil, &ctrait.Error{
		Internal: ctrait.ErrNotFound,
		Detail:   fmt.Sprintf("%s oprator not support", op.Command),
	}
}

func init() {
	executorBuilder = make(map[string]func(objs ObjectStore, op trait.Operator) (Executor, trait.Error))
}
