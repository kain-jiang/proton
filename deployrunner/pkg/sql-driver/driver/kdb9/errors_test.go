package store

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"taskrunner/test"
	testpkg "taskrunner/test/pkg"
	"taskrunner/trait"

	"github.com/AISHU-Technology/proton-rds-sdk-go/driver/kingbase/gokb"
)

func TestInitTable(t *testing.T) {
	// warn, KDB9 cfg0.User operator don't support, this use case shuold run withou cfg0.User operate
	cfg0 := testpkg.GetTestMysqlWithType("KDB9")
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
	defer func() {
		_ = s.DeleteDatabase(ctx, cfg0.DBName)
	}()
	err = s.CreateDatabase(ctx, cfg0.DBName)
	tt.AssertNil(err)
	err = s.CreateDatabase(ctx, cfg0.DBName)
	tt.AssertNil(err)

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
	// defer func() {
	// 	_ = s.DeleteUser(ctx, cfg.User)
	// }()

	// err = s.GrantUserDB(ctx, cfg.User, cfg.DBName)
	// tt.AssertNil(err)

	// _, err = NewStore(ctx, *cfg)
	// if err != nil {
	// 	t.Log(err.Error())
	// 	t.FailNow()
	// }

	// s, err = NewDBOP(ctx, *cfg0)
	// tt.AssertNil(err)
	// err = s.CreateUser(ctx, cfg.User, cfg.Password)
	// tt.AssertNil(err)

	// cfg.Password = ""
	// err = s.CreateUser(ctx, cfg.User, cfg.Password)
	// tt.AssertNil(err)

	// err = s.GrantUserDB(ctx, cfg.User, cfg.DBName)
	// tt.AssertNil(err)

	// _, err = NewDBOP(ctx, *cfg)
	// tt.AssertNil(err)

	// ss, err := NewStore(ctx, *cfg)
	// tt.AssertNil(err)

	// _, err = ss.ListSystemInfo(ctx, -1, -1)
	// tt.AssertNil(err)

	// err = s.DeleteUser(ctx, cfg.User)
	// tt.AssertNil(err)

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
			err: &gokb.Error{
				Code: "22001",
			},
			ecode: trait.ErrParam,
		},
		{
			err: &gokb.Error{
				Code: "42P06",
			},
			ecode: trait.ErrUniqueKey,
		},
		{
			err: &gokb.Error{
				Code: "42P04",
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
			t.Errorf("testcase %d want err but get %#v", i, err)
		}
	}
}
