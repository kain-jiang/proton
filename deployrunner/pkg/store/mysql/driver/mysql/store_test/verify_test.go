package store_test

import (
	"context"
	"testing"

	"taskrunner/test"
)

func TestVerify(t *testing.T) {
	// 当前deployrunner只提供更新验证的查询接口，数据验证功能由data-model提供，模块功能验证由模块化服务对应的验证组件提供
	s := getTestStore(t)
	defer cleanTestStore(s)
	tt := test.TestingT{T: t}
	ctx := context.Background()
	jid := 1
	_, err := s.GetVerifyRecord(ctx, jid)
	tt.AssertNil(err)

	did := 1
	offset := 0
	pageSize := 0
	_, err = s.GetDataTestEntries(ctx, did, offset, pageSize)
	tt.AssertNil(err)
	_, err = s.CountDataTestEntries(ctx, did)
	tt.AssertNil(err)

	fid := 1
	_, err = s.GetFunctionTestEntries(ctx, fid, offset, pageSize)
	tt.AssertNil(err)
	_, err = s.CountFunctionTestEntries(ctx, fid)
	tt.AssertNil(err)
}
