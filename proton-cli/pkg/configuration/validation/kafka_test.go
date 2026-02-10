package validation

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateKafka(t *testing.T) {
	type args struct {
		k           *configuration.Kafka
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
				k: &configuration.Kafka{
					Hosts: []string{
						"node-0",
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
				k: &configuration.Kafka{
					StorageClassName: "standard",
				},
			},
		},
		{
			name: "invalid-host-undefined",
			args: args{
				k: &configuration.Kafka{
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
			name: "invalid-data-path-missing",
			args: args{
				k: &configuration.Kafka{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateKafka(tt.args.k, tt.args.nodeNameSet, tt.args.fldPath)
			for _, err := range errs {
				t.Log(err)
			}

			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateKafka() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateKafkaUpdate(t *testing.T) {
	type args struct {
		o       *configuration.Kafka
		n       *configuration.Kafka
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
				o: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
				n: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
			name: "valid-version",
			args: args{
				o: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
				n: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
			name: "valid-env",
			args: args{
				o: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
				n: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms8G -Xmx8G",
					},
					Resources: corev1.ResourceRequirements{
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
				o: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms8G -Xmx8G",
					},
					Resources: corev1.ResourceRequirements{
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
				n: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms8G -Xmx8G",
					},
					Resources: corev1.ResourceRequirements{
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
				o: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
				n: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/1",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
			name: "invalid-old-host-not-front",
			args: args{
				o: &configuration.Kafka{
					Hosts: []string{
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
				n: &configuration.Kafka{
					Hosts: []string{
						"node-0",
						"node-1",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
			name: "valid-hosts-expansion",
			args: args{
				o: &configuration.Kafka{
					Hosts: []string{
						"node-1",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
				n: &configuration.Kafka{
					Hosts: []string{
						"node-1",
						"node-0",
						"node-2",
					},
					Data_path: "/data/path/0",
					Env: map[string]string{
						"KAFKA_HEAP_OPTS": "-Xms9G -Xmx9G",
					},
					Resources: corev1.ResourceRequirements{
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
			name: "invalid storage class name is changed",
			args: args{
				o: &configuration.Kafka{
					StorageClassName: "storage-class-name-old",
				},
				n: &configuration.Kafka{
					StorageClassName: "storage-class-name-new",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateKafkaUpdate(tt.args.o, tt.args.n, tt.args.fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				for i, err := range errList {
					t.Errorf("ValidateKafkaUpdate() errList[%d] = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}
