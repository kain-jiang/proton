package utils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"taskrunner/test"
	"taskrunner/trait"
)

func TestRetryN(t *testing.T) {
	tt := test.TestingT{T: t}
	count := 0
	ctx := context.Background()
	interval := 1 * time.Millisecond
	err := &trait.Error{
		Err:      fmt.Errorf("test n count"),
		Internal: trait.ECNULL,
	}
	_ = RetryN(ctx, func() (bool, *trait.Error) {
		count++
		return true, err
	}, 2, interval)
	tt.Assert(2, count)

	ctx0, cancel := context.WithCancel(ctx)
	cancel()
	count = 0
	_ = RetryN(ctx0, func() (bool, *trait.Error) {
		count++
		return true, err
	}, 2, interval)
	tt.Assert(1, count)
}
