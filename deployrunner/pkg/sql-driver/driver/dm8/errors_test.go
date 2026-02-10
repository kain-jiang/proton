package store

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"

	"taskrunner/test"
	testpkg "taskrunner/test/pkg"
	"taskrunner/trait"

	"gitee.com/chunanyong/dm"
)

// func getTestStore(t *testing.T) *store.Store {
// 	cfg := testpkg.GetTestMysqlWithType("DM8")
// 	tt := test.TestingT{T: t}
// 	if cfg == nil {
// 		t.SkipNow()
// 	}

// 	db, err := NewDBOP(context.Background(), *cfg)
// 	tt.AssertNil(err)
// 	err = db.CreateDatabase(context.Background(), cfg.DBName)
// 	tt.AssertNil(err)
// 	s, err := NewStore(context.TODO(), *cfg)
// 	tt.AssertNil(err)

// 	return s
// }

// func cleanTestStore() {
// 	cfg := testpkg.GetTestMysqlWithType("DM8")
// 	if cfg == nil {
// 		panic("get conf fail")
// 	}
// 	db, err := NewDBOP(context.Background(), *cfg)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err := db.DeleteDatabase(context.Background(), cfg.DBName); err != nil {
// 		panic(err)
// 	}
// }

func TestInsertChart(t *testing.T) {
	t.SkipNow()
	cfg0 := testpkg.GetTestMysqlWithType("DM8")
	if cfg0 == nil {
		t.SkipNow()
	}
	ctx := context.Background()
	tt := test.TestingT{T: t}
	s, err := NewDBOP(ctx, *cfg0)
	tt.AssertNil(err)
	defer s.Close()
	err = s.CreateDatabase(ctx, cfg0.DBName)
	tt.AssertNil(err)
	defer func() {
		err = s.DeleteDatabase(ctx, cfg0.DBName)
		tt.AssertNil(err)
	}()

	_, _ = s.ExecContext(ctx, `drop table test;`)
	_, err = s.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS test (
		aaa VARCHAR(10)
	) ;`)
	tt.AssertNil(err)

	buf := bytes.NewBuffer(nil)
	for i := 1; i < 42; i++ {
		buf.WriteString("a")
	}
	_, err = s.ExecContext(ctx, "INSERT INTO test (aaa) VALUES (?);", buf.String())
	if err != nil {
		tt.AssertNil(err)
		// t.Errorf("i: %d, err: %s", i, err.Error())
		return
	}

	row := s.QueryRowContext(ctx, "select aaa from test")
	str := ""
	err = row.Scan(&str)
	tt.AssertNil(err)
	t.Log(str)
	t.Log(len(str))
	tt.AssertNil(s.Close())
	t.Log("")
}

func TestInitTable(t *testing.T) {
	// warn, dm8 cfg0.User operator don't support, this use case shuold run withou cfg0.User operate
	cfg0 := testpkg.GetTestMysqlWithType("DM8")
	if cfg0 == nil {
		t.SkipNow()
	}
	cfg := testpkg.GetTestMysql()
	cfg.Password = "FAKE_PASSWORD"
	cfg.User = "FAKE_USERNAME"
	ctx := context.Background()
	tt := test.TestingT{T: t}
	s, err := NewDBOP(ctx, *cfg0)
	tt.AssertNil(err)
	defer s.Close()
	err = s.CreateDatabase(ctx, cfg0.DBName)
	tt.AssertNil(err)
	defer func() {
		_ = s.DeleteDatabase(ctx, cfg0.DBName)
	}()

	err = s.CreateUser(ctx, cfg.User, cfg.Password)
	tt.AssertNil(err)
	defer func() {
		_ = s.DeleteUser(ctx, cfg.User)
	}()

	err = s.GrantUserDB(ctx, cfg.User, cfg.DBName)
	tt.AssertNil(err)

	_, err = NewStore(ctx, *cfg0)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	s, err = NewDBOP(ctx, *cfg0)
	tt.AssertNil(err)
	err = s.CreateUser(ctx, cfg.User, cfg.Password)
	tt.AssertNil(err)

	cfg.Password = "FAKE_PASSWORD_ALT"
	err = s.CreateUser(ctx, cfg.User, cfg.Password)
	tt.AssertNil(err)

	err = s.GrantUserDB(ctx, cfg.User, cfg.DBName)
	tt.AssertNil(err)

	// _, err = NewDBOP(ctx, *cfg)
	// tt.AssertNil(err)

	// ss, err := NewStore(ctx, *cfg)
	// tt.AssertNil(err)

	// _, err = ss.ListSystemInfo(ctx, -1, -1)
	// tt.AssertNil(err)

	err = s.DeleteUser(ctx, cfg.User)
	tt.AssertNil(err)

	err = s.DeleteDatabase(ctx, cfg.DBName)
	tt.AssertNil(err)
}

func TestErrorWrrapper(t *testing.T) {
	tt := test.TestingT{T: t}
	ts := []struct {
		wantNil bool
		err     error
		ecode   int
	}{
		{
			err:     nil,
			wantNil: true,
		},
		{
			err: &dm.DmError{
				ErrCode: -5403,
			},
			ecode: trait.ErrParam,
		},
		{
			err: &dm.DmError{
				ErrCode: -6169,
			},
			ecode: trait.ErrParam,
		},
		{
			err: &dm.DmError{
				ErrCode: -6602,
			},
			ecode: trait.ErrUniqueKey,
		},
		{
			err: &dm.DmError{
				ErrCode: -6625,
			},
			ecode: trait.ErrUniqueKey,
		},
		{
			err:   sql.ErrNoRows,
			ecode: trait.ErrNotFound,
		},
		{
			err:   fmt.Errorf("test unknow"),
			ecode: trait.ECSQLUnknow,
		},
		{
			err:   &trait.Error{Internal: trait.ECSQLUnknow, Detail: "unknow error for sql database"},
			ecode: trait.ECSQLUnknow,
		},
	}

	for i, tc := range ts {
		err := writerErrorWrraper(tc.err)
		if tc.wantNil {
			tt.AssertNil(err)
		} else if !trait.IsInternalError(err, tc.ecode) {
			t.Errorf("testcase %d want err but get nil", i)
		}
	}
}
