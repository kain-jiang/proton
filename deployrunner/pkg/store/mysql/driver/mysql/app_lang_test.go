package store_test

import (
	"context"
	"testing"

	"taskrunner/test"
)

func TestAppLang(t *testing.T) {
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()

	lang := "zh-cn"
	aname := "test"
	alias := "test"
	tt.AssertNil(s.InsertAppLang(ctx, lang, aname, alias, aname))
	alias = "测试"
	tt.AssertNil(s.InsertAppLang(ctx, lang, aname, alias, aname))
	a, err := s.GetAppLang(ctx, lang, aname, aname)
	tt.AssertNil(err)
	tt.Assert(a, alias)

	// tt.Assert("", s.GetAname(lang, alias, aname))
}
