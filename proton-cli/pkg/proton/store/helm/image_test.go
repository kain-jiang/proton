package helm

import (
	"testing"

	"github.com/go-test/deep"
)

func Test_imageFor(t *testing.T) {
	type args struct {
		registry string
	}
	tests := []struct {
		name string
		args args
		want Image
	}{
		{
			name: "example",
			args: args{
				registry: "registry.example.org",
			},
			want: Image{
				Registry: "registry.example.org",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := imageFor(tt.args.registry)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("imageFor() got != want: %v", d)
			}
		})
	}
}
