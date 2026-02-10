package operator

import (
	"context"
	"encoding/json"
	"fmt"

	etrait "taskrunner/error/codes/trait"
	driver "taskrunner/pkg/sql-driver"
	"taskrunner/pkg/store/mysql/upgrade/store"
	"taskrunner/pkg/store/mysql/upgrade/trait"
	ctrait "taskrunner/trait"
)

// SQLExecuteCommand command const
const SQLExecuteCommand = "execute"

func init() {
	executorBuilder[SQLExecuteCommand] = newSQLExeute
}

// SQLExeuteArgs simple sql args
type SQLExeuteArgs struct {
	Statements  []string `json:"statements"`
	Transaction bool     `json:"transaction"`
	DBObj       string   `json:"db"`
}

// SQLExecute simple sql operator
type SQLExecute struct {
	conn driver.DBConn
	SQLExeuteArgs
	rawOPDefine trait.Operator
}

func newSQLExeute(objs ObjectStore, op trait.Operator) (Executor, trait.Error) {
	e := &SQLExecute{
		rawOPDefine: op,
	}
	if err := json.Unmarshal(op.Args, &e.SQLExeuteArgs); err != nil {
		return nil, &ctrait.Error{
			Internal: ctrait.ErrParam,
			Err:      err,
		}
	}
	if e.SQLExeuteArgs.DBObj == "" {
		e.SQLExeuteArgs.DBObj = "default"
	}
	db := objs[e.SQLExeuteArgs.DBObj]
	if db == nil {
		return nil, &ctrait.Error{
			Internal: ctrait.ErrNotFound,
			Detail:   fmt.Sprintf("objects %s not found", e.SQLExeuteArgs.DBObj),
		}
	}
	conn, ok := db.(driver.DBConn)
	if !ok {
		return nil, &ctrait.Error{
			Internal: ctrait.ErrNotFound,
			Detail:   fmt.Sprintf("objects %s is'n sql database operator", e.SQLExeuteArgs.DBObj),
		}
	}
	e.conn = conn
	return e, nil
}

// Execute imply operator executor
func (e *SQLExecute) Execute(ctx context.Context, w *trait.WorkEnv, s store.Store, pm trait.PlanProcess) (err trait.Error) {
	var conn driver.CursorConn
	conn = e.conn
	var tx driver.Transaction
	log := w.Log
	if e.Transaction {
		tx, err = e.conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		conn = tx
	}

	for j, i := range e.Statements {
		_, err = conn.ExecContext(ctx, i)
		if ctrait.IsInternalError(err, etrait.ECColumnExists) {
			err = nil
		} else if ctrait.IsInternalError(err, etrait.ECPriKeyExists) {
			err = nil
		} else if err != nil {
			log.Errorf(
				"执行%s计划%d的第%d个算子的第%d个语句失败",
				pm.ServiceName, pm.Order, pm.Op.OrderID, j)
			log.Trace(i)
			if tx != nil {
				if rerr := tx.Rollback(); rerr != nil {
					return rerr
				}
			}

			pm.Op.Status = ctrait.AppFailStatus
			if rerr := s.Record(ctx, pm); rerr != nil {
				return rerr
			}
			return err
		}
	}

	if tx != nil {
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return err
}
