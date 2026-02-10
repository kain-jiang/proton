package operator

import (
	"context"
	"encoding/json"
	"testing"

	"taskrunner/pkg/sql-driver/driver"
	"taskrunner/pkg/store/mysql/upgrade/store"
	"taskrunner/pkg/store/mysql/upgrade/trait"
	"taskrunner/test"
	testpkg "taskrunner/test/pkg"
	ctrait "taskrunner/trait"

	"github.com/sirupsen/logrus"
)

var _DataBaseType = "MARIADB"

func getTestStore(t *testing.T) (store.Store, driver.DBConn) {
	ctx := context.Background()
	cfg := testpkg.GetTestMysqlWithType(_DataBaseType)
	tt := test.TestingT{T: t}
	if cfg == nil {
		t.SkipNow()
	}

	op, err := driver.Factory.NewDBOP(ctx, *cfg)
	tt.AssertNil(err)
	tt.AssertNil(op.CreateDatabase(ctx, cfg.DBName))

	db, err := store.NewStore(context.Background(), *cfg)
	tt.AssertNil(err)
	conn, err := driver.Factory.NewDBConn(ctx, *cfg)
	tt.AssertNil(err)

	return db, conn
}

func cleanTestStore(t *testing.T) {
	ctx := context.Background()
	cfg := testpkg.GetTestMysqlWithType(_DataBaseType)
	tt := test.TestingT{T: t}
	if cfg == nil {
		t.SkipNow()
	}

	op, err := driver.Factory.NewDBOP(ctx, *cfg)
	tt.AssertNil(err)
	tt.AssertNil(op.DeleteDatabase(ctx, cfg.DBName))
}

func TestExecute(t *testing.T) {
	ctx := context.Background()
	tt := test.TestingT{T: t}
	defer cleanTestStore(t)
	s, db := getTestStore(t)
	objs := map[string]any{
		"default": db,
	}
	w := &trait.WorkEnv{
		Log: logrus.New(),
	}

	ts := []struct {
		wantErr bool
		args    SQLExeuteArgs
	}{
		{
			wantErr: false,
			args: SQLExeuteArgs{
				Transaction: true,
				Statements: []string{
					"SELECT 1 FROM DUAL;",
					"SELECT 2 FROM DUAL;",
				},
			},
		},
		{
			wantErr: true,
			args: SQLExeuteArgs{
				Transaction: true,
				Statements: []string{
					"qwe DUAL;",
				},
			},
		},
	}

	for i, tc := range ts {
		bs, rerr := json.Marshal(tc.args)
		tt.AssertNil(rerr)
		op := trait.Operator{
			OperatorMeta: trait.OperatorMeta{
				Command: SQLExecuteCommand,
			},
			Args: bs,
		}
		e, err := NewExecutor(objs, op)
		tt.AssertNil(err)
		pp := trait.PlanProcess{}
		err = e.Execute(ctx, w, s, pp)
		if !tc.wantErr {
			tt.AssertNil(err)
		} else if err == nil {
			t.Errorf("testcase %d want error, but success", i)
			t.FailNow()
		}
	}

	_, err := newSQLExeute(objs, trait.Operator{
		OperatorMeta: trait.OperatorMeta{
			Command: SQLExecuteCommand,
		},
		Args: nil,
	})
	tt.AssertError(ctrait.ErrParam, err)

	_, err = newSQLExeute(objs, trait.Operator{
		OperatorMeta: trait.OperatorMeta{
			Command: SQLExecuteCommand,
		},
		Args: []byte(`{"db": "qwe"}`),
	})
	tt.AssertError(ctrait.ErrNotFound, err)
}
