package helm

import (
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/api/resource"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func Test_storageFor(t *testing.T) {

	type args struct {
		spec *configuration.PackageStore
	}
	tests := []struct {
		name string
		args args
		want Storage
	}{
		{
			name: "bare",
			args: args{
				spec: &configuration.PackageStore{
					Hosts: []string{"node-0", "node-1"},
					Storage: configuration.PackageStoreStorage{
						Capacity: resource.NewQuantity(128<<20, resource.BinarySI),
						Path:     "/var/lib/proton-package-store",
					},
				},
			},
			want: Storage{
				Capacity: resource.MustParse("128Mi"),
				Local:    localFor([]string{"node-0", "node-1"}, "/var/lib/proton-package-store"),
			},
		},
		{
			name: "host",
			args: args{
				spec: &configuration.PackageStore{
					Storage: configuration.PackageStoreStorage{
						StorageClassName: "csi-disk",
						Capacity:         resource.NewQuantity(16<<30, resource.BinarySI),
					},
				},
			},
			want: Storage{
				StorageClassName: "csi-disk",
				Capacity:         resource.MustParse("16Gi"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := storageFor(tt.args.spec)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("storageFor() got != want: %v", d)
			}
		})
	}
}

func Test_localFor(t *testing.T) {
	type args struct {
		hosts []string
		path  string
	}
	tests := []struct {
		name string
		args args
		want map[string]Local
	}{
		{
			name: "single",
			args: args{
				hosts: []string{"node-0"},
				path:  "/var/lib/proton-package-store",
			},
			want: map[string]Local{
				"0": {
					Host: "node-0",
					Path: "/var/lib/proton-package-store",
				},
			},
		},
		{
			name: "multi",
			args: args{
				hosts: []string{"node-0", "node-1", "node-2"},
				path:  "/var/lib/proton-package-store",
			},
			want: map[string]Local{
				"0": {
					Host: "node-0",
					Path: "/var/lib/proton-package-store",
				},
				"1": {
					Host: "node-1",
					Path: "/var/lib/proton-package-store",
				},
				"2": {
					Host: "node-2",
					Path: "/var/lib/proton-package-store",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := localFor(tt.args.hosts, tt.args.path)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("localFor() got != want: %v", d)
			}
		})
	}
}
