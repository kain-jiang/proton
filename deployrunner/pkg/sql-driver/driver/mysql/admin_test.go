package store_test

import (
	"context"
	"strings"
	"testing"

	"taskrunner/test"

	testpkg "taskrunner/test/pkg"
)

func TestInitTable(t *testing.T) {
	// warn, dm8 cfg0.User operator don't support, this use case shuold run withou cfg0.User operate
	cfg0 := testpkg.GetTestMysqlWithType(_DataBaseType)
	if cfg0 == nil {
		t.SkipNow()
	}
	cfg := testpkg.GetTestMysqlWithType(_DataBaseType)
	cfg.Password = "deployrunner"
	cfg.User = "deployrunner"
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

	if strings.ToLower(cfg0.Type) == "dm8" {
		return
	}

	err = s.CreateUser(ctx, cfg.User, cfg.Password)
	tt.AssertNil(err)
	defer func() {
		_ = s.DeleteUser(ctx, cfg.User)
	}()

	err = s.GrantUserDB(ctx, cfg.User, cfg.DBName)
	tt.AssertNil(err)

	_, err = NewStore(ctx, *cfg)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	s, err = NewDBOP(ctx, *cfg0)
	tt.AssertNil(err)
	err = s.CreateUser(ctx, cfg.User, cfg.Password)
	tt.AssertNil(err)

	cfg.Password = ""
	err = s.CreateUser(ctx, cfg.User, cfg.Password)
	tt.AssertNil(err)

	err = s.GrantUserDB(ctx, cfg.User, cfg.DBName)
	tt.AssertNil(err)

	_, err = NewDBOP(ctx, *cfg)
	tt.AssertNil(err)

	// ss, err := NewStore(ctx, *cfg)
	// tt.AssertNil(err)

	// _, err = ss.ListSystemInfo(ctx, -1, -1)
	// tt.AssertNil(err)

	err = s.DeleteUser(ctx, cfg.User)
	tt.AssertNil(err)

	err = s.DeleteDatabase(ctx, cfg.DBName)
	tt.AssertNil(err)
}
