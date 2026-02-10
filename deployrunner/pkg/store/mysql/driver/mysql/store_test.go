package store_test

import (
	"context"
	"fmt"
	"testing"

	"taskrunner/pkg/component/resources"
	"taskrunner/test"
	testpkg "taskrunner/test/pkg"
	"taskrunner/trait"

	"github.com/mohae/deepcopy"
	// . "taskrunner/pkg/sql-driver/driver/mysql"
)

func getTestStore(t *testing.T) Store {
	cfg := testpkg.GetTestMysqlWithType(_DataBaseType)
	tt := test.TestingT{T: t}
	if cfg == nil {
		t.SkipNow()
	}
	ctx := context.Background()
	db, err := NewDBOP(context.Background(), *cfg)
	tt.AssertNil(err)
	err = db.CreateDatabase(context.Background(), cfg.DBName)
	tt.AssertNil(err)
	s, err := NewStore(context.TODO(), *cfg)
	tt.AssertNil(err)
	tt.AssertNil(s.InitTablesFromDir(ctx, _testInitSqlDir))

	return s
}

func cleanTestStore(s trait.Store) {
	s.Close()
	cfg := testpkg.GetTestMysqlWithType(_DataBaseType)
	if cfg == nil {
		panic(fmt.Sprintf("get %s config fail", _DataBaseType))
	}
	db, err := NewDBOP(context.Background(), *deepcopy.Copy(cfg).(*resources.RDS))
	if err != nil {
		panic(err)
	}

	if err := db.DeleteDatabase(context.Background(), cfg.DBName); err != nil {
		panic(err)
	}
}

func TestConnectDB(t *testing.T) {
	cfg := testpkg.GetTestMysqlWithType(_DataBaseType)
	if cfg == nil {
		t.SkipNow()
	}
	tt := test.TestingT{T: t}

	ctx := context.Background()
	// cfg0 := deepcopy.Copy(cfg).(resources.RDS)
	// cfg0.Host = ""
	// _, err := newDB(ctx, cfg0)
	// errStr := fmt.Sprintf("default addr for network '" + cfg0.Net + "' unknown")
	// tt.Assert(errStr, err.Error())

	// _, err = NewStore(ctx, cfg0)
	// tt.Assert(errStr, err.Error())

	db, err := NewDBOP(ctx, *cfg)
	tt.AssertNil(err)
	err = db.CreateDatabase(ctx, cfg.DBName)
	tt.AssertNil(err)

	s, err := NewStore(ctx, *cfg)
	tt.AssertNil(err)
	tt.AssertNil(s.InitTablesFromDir(ctx, _testInitSqlDir))

	_, err = NewStore(ctx, *cfg)
	tt.AssertNil(err)
	defer cleanTestStore(s)

	tx, err := s.Begin(ctx)
	tt.AssertNil(err)

	tt.AssertNil(tx.Commit())

	tx, err = s.Begin(ctx)
	tt.AssertNil(err)
	tt.AssertNil(tx.Rollback())

	// err = s.createDatabase(ctx)
	// tt.Assert(t, errStr, errStr)

	// err = s.deleteDatabase(ctx)
	// tt.Assert(t, errStr, errStr)

	// s.cfg = cfg
	// s.createDatabase(ctx)
	// err = s.createDatabase(ctx)
	// tt.Assert(t, nil, err)

	// err = s.deleteDatabase(ctx)
	// tt.Assert(t, nil, err)
}
