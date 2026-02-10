package check

import (
	"errors"
	"testing"
)

func TestErrDirectoryNotEmpty_Error(t *testing.T) {
	tests := []struct {
		name string
		e    *ErrDirectoryNotEmpty
		want string
	}{
		{
			name: "example",
			e:    &ErrDirectoryNotEmpty{Path: "/var/lib/something"},
			want: "/var/lib/something is not empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ErrDirectoryNotEmpty.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrDirectoryNotEmpty_Is(t *testing.T) {
	type fields struct {
		Path string
	}
	type args struct {
		target error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ErrNotEmpty",
			args: args{target: ErrNotEmpty},
			want: true,
		},
		{
			name: "ErrOther",
			args: args{target: errors.New("other")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ErrDirectoryNotEmpty{
				Path: tt.fields.Path,
			}
			if got := e.Is(tt.args.target); got != tt.want {
				t.Errorf("ErrDirectoryNotEmpty.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}
