package system

import (
	"testing"

	"taskrunner/test"
)

func TestNumToLetter(t *testing.T) {
	tt := test.TestingT{T: t}
	ts := []struct {
		num  int
		want string
	}{
		{
			num:  0,
			want: "a",
		},
		{
			num:  1,
			want: "b",
		},
		{
			num:  25,
			want: "z",
		},
		{
			num:  26,
			want: "ab",
		},
	}

	for _, i := range ts {
		get := numToLetter(i.num)
		tt.Assert(i.want, get)
	}
}
