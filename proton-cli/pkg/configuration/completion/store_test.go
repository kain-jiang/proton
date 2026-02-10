package completion

import (
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/store"
)

func TestCompletePackageStore(t *testing.T) {
	tests := []struct {
		name      string
		cfg, want *configuration.PackageStore
	}{
		{
			name: "bared",
			cfg: &configuration.PackageStore{
				Hosts: []string{"node-0", "node-1"},
			},
			want: &configuration.PackageStore{
				Hosts:    []string{"node-0", "node-1"},
				Replicas: ptr.To(2),
				Storage: configuration.PackageStoreStorage{
					Capacity: resource.NewQuantity(store.DefaultStorageCapacity, resource.BinarySI),
					Path:     store.DefaultStoragePath,
				},
			},
		},
		{
			name: "hosted",
			cfg: &configuration.PackageStore{
				Storage: configuration.PackageStoreStorage{
					StorageClassName: "csi-disk",
				},
			},
			want: &configuration.PackageStore{
				Replicas: ptr.To(store.DefaultReplicas),
				Storage: configuration.PackageStoreStorage{
					StorageClassName: "csi-disk",
					Capacity:         resource.NewQuantity(store.DefaultStorageCapacity, resource.BinarySI),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompletePackageStore(tt.cfg)
			for _, d := range deep.Equal(tt.cfg, tt.want) {
				t.Errorf("CompletePackageStore() got != want: %v", d)
			}
		})
	}
}

func TestCompleteBaredPackageStore(t *testing.T) {
	tests := []struct {
		name      string
		cfg, want *configuration.PackageStore
	}{
		{
			name: "replicas",
			cfg: &configuration.PackageStore{
				Hosts: []string{"node-0", "node-1"},
				Storage: configuration.PackageStoreStorage{
					Capacity: resource.NewQuantity(1024, resource.BinarySI),
					Path:     "/var/lib/proton-package-store",
				},
			},
			want: &configuration.PackageStore{
				Hosts:    []string{"node-0", "node-1"},
				Replicas: ptr.To(2),
				Storage: configuration.PackageStoreStorage{
					Capacity: resource.NewQuantity(1024, resource.BinarySI),
					Path:     "/var/lib/proton-package-store",
				},
			},
		},
		{
			name: "storage.capacity",
			cfg: &configuration.PackageStore{
				Hosts:    []string{"node-0", "node-1"},
				Replicas: ptr.To(2),
				Storage: configuration.PackageStoreStorage{
					Path: "/var/lib/proton-package-store",
				},
			},
			want: &configuration.PackageStore{
				Hosts:    []string{"node-0", "node-1"},
				Replicas: ptr.To(2),
				Storage: configuration.PackageStoreStorage{
					Capacity: resource.NewQuantity(store.DefaultStorageCapacity, resource.BinarySI),
					Path:     "/var/lib/proton-package-store",
				},
			},
		},
		{
			name: "storage.path",
			cfg: &configuration.PackageStore{
				Hosts:    []string{"node-0", "node-1"},
				Replicas: ptr.To(2),
				Storage: configuration.PackageStoreStorage{
					Capacity: resource.NewQuantity(1024, resource.BinarySI),
				},
			},
			want: &configuration.PackageStore{
				Hosts:    []string{"node-0", "node-1"},
				Replicas: ptr.To(2),
				Storage: configuration.PackageStoreStorage{
					Capacity: resource.NewQuantity(1024, resource.BinarySI),
					Path:     store.DefaultStoragePath,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompleteBaredPackageStore(tt.cfg)
			for _, d := range deep.Equal(tt.cfg, tt.want) {
				t.Errorf("CompleteBaredPackageStore() got != want: %v", d)
			}
		})
	}
}

func TestCompleteHostedPackageStore(t *testing.T) {
	tests := []struct {
		name      string
		cfg, want *configuration.PackageStore
	}{
		{
			name: "replicas",
			cfg: &configuration.PackageStore{
				Storage: configuration.PackageStoreStorage{
					StorageClassName: "csi-disk",
					Capacity:         resource.NewQuantity(12450<<10, resource.BinarySI),
				},
			},
			want: &configuration.PackageStore{
				Replicas: ptr.To(store.DefaultReplicas),
				Storage: configuration.PackageStoreStorage{
					StorageClassName: "csi-disk",
					Capacity:         resource.NewQuantity(12450<<10, resource.BinarySI),
				},
			},
		},
		{
			name: "storage.capacity",
			cfg: &configuration.PackageStore{
				Replicas: ptr.To(12450),
				Storage: configuration.PackageStoreStorage{
					StorageClassName: "csi-disk",
				},
			},
			want: &configuration.PackageStore{
				Replicas: ptr.To(12450),
				Storage: configuration.PackageStoreStorage{
					StorageClassName: "csi-disk",
					Capacity:         resource.NewQuantity(store.DefaultStorageCapacity, resource.BinarySI),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompleteHostedPackageStore(tt.cfg)
			for _, d := range deep.Equal(tt.cfg, tt.want) {
				t.Errorf("CompleteHostedPackageStore() got != want: %v", d)
			}
		})
	}
}
