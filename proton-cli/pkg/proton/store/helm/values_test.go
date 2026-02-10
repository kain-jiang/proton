package helm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValuesFor(t *testing.T) {
	type args struct {
		spec     *configuration.PackageStore
		registry string
		rds      *configuration.RdsInfo
		database string
	}
	tests := []struct {
		name string
		args args
		want *Values
	}{
		{
			name: "bare",
			args: args{
				spec: &configuration.PackageStore{
					Hosts:    []string{"node-0", "node-1"},
					Replicas: ptr.To(2),
					Storage: configuration.PackageStoreStorage{
						Capacity: resource.NewQuantity(16<<20, resource.BinarySI),
						Path:     "/var/lib/proton-package-store",
					},
					Resources: &corev1.ResourceRequirements{
						Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
						Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
					},
				},
				registry: "registry.aishu.cn:15000",
				rds: &configuration.RdsInfo{
					SourceType: configuration.Internal,
					RdsType:    configuration.MariaDB,
					Hosts:      "mariadb-mariadb-cluster.resource",
					Port:       3330,
					Username:   "hello",
					Password:   "world",
				},
				database: "test-db",
			},
			want: &Values{
				Image: Image{
					Registry: "registry.aishu.cn:15000",
				},
				ReplicaCount: 2,
				DepServices: DepServices{
					RDS: RDS{
						Host:     "mariadb-mariadb-cluster.resource",
						Port:     3330,
						Username: "hello",
						Password: "world",
						Database: "test-db",
					},
				},
				Storage: Storage{
					Capacity: resource.MustParse("16Mi"),
					Local: map[string]Local{
						"0": {
							Host: "node-0",
							Path: "/var/lib/proton-package-store",
						},
						"1": {
							Host: "node-1",
							Path: "/var/lib/proton-package-store",
						},
					},
				},
				Resources: &Resources{
					Store: &corev1.ResourceRequirements{
						Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
						Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
					},
				},
				Namespace: "resource",
			},
		},
		{
			name: "host",
			args: args{
				spec: &configuration.PackageStore{
					Replicas: ptr.To(3),
					Storage: configuration.PackageStoreStorage{
						StorageClassName: "csi-disk",
						Capacity:         resource.NewQuantity(8<<30, resource.BinarySI),
					},
					Resources: &corev1.ResourceRequirements{
						Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
						Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
					},
				},
				registry: "registry.example.org",
				rds: &configuration.RdsInfo{
					SourceType: configuration.External,
					RdsType:    configuration.MariaDB,
					Hosts:      "rds.example.org",
					Port:       3306,
					Username:   "hello",
					Password:   "world",
				},
				database: "test-db",
			},
			want: &Values{
				Image: Image{
					Registry: "registry.example.org",
				},
				ReplicaCount: 3,
				DepServices: DepServices{
					RDS: RDS{
						Host:     "rds.example.org",
						Port:     3306,
						Username: "hello",
						Password: "world",
						Database: "test-db",
					},
				},
				Storage: Storage{
					StorageClassName: "csi-disk",
					Capacity:         resource.MustParse("8Gi"),
				},
				Resources: &Resources{
					Store: &corev1.ResourceRequirements{
						Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
						Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
					},
				},
				Namespace: "resource",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValuesFor(tt.args.spec, tt.args.registry, tt.args.rds, tt.args.database, "resource")
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("ValuesFor() got != want: %v", d)
			}
		})
	}
}

func loadMapFromFile(t *testing.T, name string) map[string]interface{} {
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	var v Values
	if err := yaml.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	return v.ToMap()
}

func TestValues_ToMap(t *testing.T) {
	type fields struct {
		Image        Image
		ReplicaCount int
		DepServices  DepServices
		Storage      Storage
		Resources    *Resources
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "bare",
			fields: fields{
				Image: Image{
					Registry: "registry.aishu.cn:15000",
				},
				ReplicaCount: 2,
				DepServices: DepServices{
					RDS: RDS{
						Host:     "mariadb-mariadb-cluster.resource",
						Port:     3330,
						Username: "hello",
						Password: "world",
					},
				},
				Storage: Storage{
					Capacity: resource.MustParse("16Mi"),
					Local: map[string]Local{
						"0": {
							Host: "node-0",
							Path: "/var/lib/proton-package-store",
						},
						"1": {
							Host: "node-1",
							Path: "/var/lib/proton-package-store",
						},
					},
				},
				Resources: &Resources{
					Store: &corev1.ResourceRequirements{
						Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
						Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
					},
				},
			},
			want: "bare.yaml",
		},
		{
			name: "bare without resources",
			fields: fields{
				Image: Image{
					Registry: "registry.aishu.cn:15000",
				},
				ReplicaCount: 2,
				DepServices: DepServices{
					RDS: RDS{
						Host:     "mariadb-mariadb-cluster.resource",
						Port:     3330,
						Username: "hello",
						Password: "world",
					},
				},
				Storage: Storage{
					Capacity: resource.MustParse("16Mi"),
					Local: map[string]Local{
						"0": {
							Host: "node-0",
							Path: "/var/lib/proton-package-store",
						},
						"1": {
							Host: "node-1",
							Path: "/var/lib/proton-package-store",
						},
					},
				},
			},
			want: "bare-without-resources.yaml",
		},
		{
			name: "host",
			fields: fields{
				Image: Image{
					Registry: "registry.example.org",
				},
				ReplicaCount: 2,
				DepServices: DepServices{
					RDS: RDS{
						Host:     "rds.example.org",
						Port:     3306,
						Username: "hello",
						Password: "world",
					},
				},
				Storage: Storage{
					StorageClassName: "csi-disk",
					Capacity:         resource.MustParse("16Mi"),
				},
				Resources: &Resources{
					Store: &corev1.ResourceRequirements{
						Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
						Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
					},
				},
			},
			want: "host.yaml",
		},
		{
			name: "host without resources",
			fields: fields{
				Image: Image{
					Registry: "registry.example.org",
				},
				ReplicaCount: 2,
				DepServices: DepServices{
					RDS: RDS{
						Host:     "rds.example.org",
						Port:     3306,
						Username: "hello",
						Password: "world",
					},
				},
				Storage: Storage{
					StorageClassName: "csi-disk",
					Capacity:         resource.MustParse("16Mi"),
				},
				Resources: &Resources{
					Store: &corev1.ResourceRequirements{
						Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
						Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
					},
				},
			},
			want: "host-without-resources.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Values{
				Image:        tt.fields.Image,
				ReplicaCount: tt.fields.ReplicaCount,
				DepServices:  tt.fields.DepServices,
				Storage:      tt.fields.Storage,
				Resources:    tt.fields.Resources,
			}
			for _, d := range deep.Equal(v.ToMap(), loadMapFromFile(t, filepath.Join("testdata", tt.want))) {
				t.Errorf("Values.ToMap() got != want: %v", d)
			}
		})
	}
}
