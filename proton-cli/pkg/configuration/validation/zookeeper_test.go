package validation

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateZooKeeper(t *testing.T) {
	type args struct {
		k           *configuration.ZooKeeper
		nodeNameSet sets.Set[string]
		fldPath     *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid-local-data-path",
			args: args{
				k: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
		},
		{
			name: "valid-storage-class",
			args: args{
				k: &configuration.ZooKeeper{
					StorageClassName: "standard",
				},
			},
		},
		{
			name: "invalid-host-undefined",
			args: args{
				k: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-x",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-too-many-hosts",
			args: args{
				k: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
						"node-3",
					},
					Data_path: "/data/path",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-data-path-missing",
			args: args{
				k: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
					},
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-both-storage-class-and-hosts",
			args: args{
				k: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
					},
					StorageClassName: "standard",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-both-storage-class-and-data-path",
			args: args{
				k: &configuration.ZooKeeper{
					Data_path:        "/data/path",
					StorageClassName: "standard",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateZooKeeper(tt.args.k, tt.args.nodeNameSet, tt.args.fldPath)
			for _, err := range errs {
				t.Log(err)
			}
			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateZooKeeper() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateZooKeeperUpdate(t *testing.T) {
	type args struct {
		o       *configuration.ZooKeeper
		n       *configuration.ZooKeeper
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid-none",
			args: args{
				o: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				n: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
		{
			name: "valid-add-host-sort",
			args: args{
				o: &configuration.ZooKeeper{
					Hosts: []string{
						"node-1",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				n: &configuration.ZooKeeper{
					Hosts: []string{
						"node-1",
						"node-0",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
		{
			name: "invalid-oldhost-not-front",
			args: args{
				o: &configuration.ZooKeeper{
					Hosts: []string{
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				n: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid-env",
			args: args{
				o: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				n: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms8G -Xmx8G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
		},
		{
			name: "valid-resources",
			args: args{
				o: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms8G -Xmx8G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				n: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms8G -Xmx8G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("101m"),
							corev1.ResourceMemory: resource.MustParse("129Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("99m"),
							corev1.ResourceMemory: resource.MustParse("127Mi"),
						},
					},
				},
			},
		},
		{
			name: "invalid-data-path",
			args: args{
				o: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
				n: &configuration.ZooKeeper{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/1",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid storage class name is changed",
			args: args{
				o: &configuration.ZooKeeper{
					StorageClassName: "storage-class-old",
				},
				n: &configuration.ZooKeeper{
					StorageClassName: "storage-class-new",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateZooKeeperUpdate(tt.args.o, tt.args.n, tt.args.fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				t.Errorf("ValidateZooKeeperUpdate() len(errList) = %v, wantErr %v", len(errList), tt.wantErr)
				for i, err := range errList {
					t.Errorf("ValidateZooKeeperUpdate() errList[%d] = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}
