package utils

import (
	"context"
	"time"

	"taskrunner/trait"
)

// RetryN run the f at lease onece.
// when f return err and want continue retry.
// run f n times at most.
// if n==0, run onece.
func RetryN(ctx context.Context, f func() (bool, *trait.Error), n int, delay time.Duration) *trait.Error {
	ok, err := f()
	for i := 1; i < n && ok; i++ {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return err
		}
		time.Sleep(delay)
		ok, err = f()
	}
	return err
}
