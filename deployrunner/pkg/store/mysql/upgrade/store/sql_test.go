package store

import (
	"context"
	"testing"

	"taskrunner/pkg/sql-driver/driver"
	"taskrunner/pkg/store/mysql/upgrade/trait"
	"taskrunner/test"
	testpkg "taskrunner/test/pkg"
)

// var _DataBaseType = "MARIADB"
var _DataBaseType = "DM8"

func getTestStore(t *testing.T) Store {
	ctx := context.Background()
	cfg := testpkg.GetTestMysqlWithType(_DataBaseType)
	tt := test.TestingT{T: t}
	if cfg == nil {
		t.SkipNow()
	}

	op, err := driver.Factory.NewDBOP(ctx, *cfg)
	tt.AssertNil(err)
	tt.AssertNil(op.CreateDatabase(ctx, cfg.DBName))

	db, err := NewStore(context.Background(), *cfg)
	tt.AssertNil(err)

	return db
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

func TestInit(t *testing.T) {
	tt := test.TestingT{T: t}
	ctx := context.Background()
	defer cleanTestStore(t)
	s := getTestStore(t)

	p := trait.PlanProcess{}
	tt.AssertNil(s.Record(ctx, p))
	p.Order = 1
	tt.AssertNil(s.Record(ctx, p))
	p0, err := s.Get(ctx, p.ServiceName, p.DateID)
	tt.AssertNil(err)
	tt.Assert(p.Order, p0.Order)

	p0.DateID = 1234
	tt.AssertNil(s.Record(ctx, *p0))

	less, err := s.Less(ctx, p.ServiceName, p0.DateID, 10, p.Stage)
	tt.AssertNil(err)
	tt.Assert(1, len(less))
	tt.Assert(p.DateID, less[0].DateID)
	last, err := s.Last(ctx, p.ServiceName, p.Stage)
	tt.AssertNil(err)
	tt.Assert(p0.DateID, last.DateID)
}
