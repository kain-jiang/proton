package trait

import (
	"context"
	"fmt"
	"runtime/debug"
	"testing"
)

func testAssert(t *testing.T, want, get interface{}) {
	if want != get {
		debug.PrintStack()
		t.Fatalf("want %#v, get %#v", want, get)
	}
}

func TestContext(t *testing.T) {
	ctx := context.Background()
	ctx0, cancel0 := WithCancelCauesContext(ctx)
	ctx1, cancel1 := WithCancelCauesContext(ctx0)
	err := &Error{
		Internal: ECExit,
		Err:      context.Canceled,
		Detail:   "test cancel engine",
	}
	// cancel transfer
	cancel0(err)
	testAssert(t, err, ctx0.Err())
	testAssert(t, err, ctx1.Err())

	err0 := &Error{
		Internal: ECExit,
		Err:      context.Canceled,
		Detail:   "test cancel engine",
	}
	// only cancel onece
	cancel1(err0)
	testAssert(t, err, ctx0.Err())
	testAssert(t, err, ctx1.Err())
}

func TestError(t *testing.T) {
	err := &Error{
		Err:      fmt.Errorf("test error"),
		Internal: 0,
		Detail:   "",
	}

	assert(t, IsInternalError(err, 0), true)
	assert(t, IsInternalError(nil, 0), false)
	assert(t, IsInternalError(err, 1), false)
	assert(t, UnwrapError(err), err)
	if UnwrapError(nil) != nil {
		t.Fatal(err)
	}
}

func assert(t *testing.T, get, want interface{}) {
	if get != want {
		debug.PrintStack()
		t.Fatalf("get %#v, want %#v", get, want)
	}
}
