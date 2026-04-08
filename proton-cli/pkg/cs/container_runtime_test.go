package cs

import (
	"reflect"
	"testing"
)

func TestRuntimePackageNames(t *testing.T) {
	tests := []struct {
		name    string
		runtime nodeContainerRuntime
		want    []string
	}{
		{
			name:    "containerd supports containerd io package",
			runtime: nodeContainerRuntimeContainerd,
			want:    []string{"containerd", "containerd.io"},
		},
		{
			name:    "docker keeps docker ce package name",
			runtime: nodeContainerRuntimeDocker,
			want:    []string{"docker-ce"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtimePackageNames(tt.runtime)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("runtimePackageNames() = %v, want %v", got, tt.want)
			}
		})
	}
}
