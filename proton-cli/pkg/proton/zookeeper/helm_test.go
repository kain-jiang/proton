package zookeeper

import (
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestHelmValuesFor(t *testing.T) {
	const (
		registry = "registry.example.org"

		nodeName0 = "node-0"

		dataPath = "/var/lib/zookeeper"
	)
	var (
		hosts = []string{nodeName0}

		helmValuesStorage = HelmValuesStorageFor(hosts, dataPath, "")

		requirements = &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("800Mi"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("200Mi"),
			},
		}
	)
	tests := []struct {
		name string
		spec *configuration.ZooKeeper
		want helm3.M
	}{
		{
			name: "with resource requirements",
			spec: &configuration.ZooKeeper{
				Hosts:     hosts,
				Data_path: dataPath,
				Env: map[string]string{
					"example-env-key": "example-env-value",
				},
				Resources: requirements,
			},
			want: helm3.M{
				"namespace":    ReleaseNamespace,
				"image":        helm3.M{"registry": registry},
				"replicaCount": 1,
				"storage":      helmValuesStorage,
				"config":       helm3.M{"zookeeperENV": helm3.M{"example-env-key": "example-env-value"}},
				"resources":    requirements,
			},
		},
		{
			name: "without resource requirements",
			spec: &configuration.ZooKeeper{
				Hosts:     []string{"node-0"},
				Data_path: "/var/lib/zookeeper",
				Env: map[string]string{
					"example-env-key": "example-env-value",
				},
			},
			want: helm3.M{
				"namespace":    ReleaseNamespace,
				"image":        helm3.M{"registry": registry},
				"replicaCount": 1,
				"storage":      helmValuesStorage,
				"config":       helm3.M{"zookeeperENV": helm3.M{"example-env-key": "example-env-value"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HelmValuesFor(tt.spec, registry)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("HelmValuesFor() got vs want: %v", d)
			}
		})
	}
}

func TestHelmValuesReplicaCountFor(t *testing.T) {
	type args struct {
		count               int
		defaultReplicaCount int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "node count",
			args: args{count: 12450, defaultReplicaCount: 10086},
			want: 12450,
		},
		{
			name: "default replica count",
			args: args{count: 0, defaultReplicaCount: 12450},
			want: 12450,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HelmValuesReplicaCountFor(tt.args.count, tt.args.defaultReplicaCount); got != tt.want {
				t.Errorf("HelmValuesReplicaCountFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHelmValuesStorageFor(t *testing.T) {
	type args struct {
		hosts            []string
		dataPath         string
		storageClassName string
	}
	tests := []struct {
		name string
		args args
		want helm3.M
	}{
		{
			name: "single local node",
			args: args{
				hosts:    []string{"node-0"},
				dataPath: "/var/lib/zookeeper",
			},
			want: helm3.M{
				"local": helm3.M{
					"0": helm3.M{
						"host": "node-0",
						"path": "/var/lib/zookeeper",
					},
				},
			},
		},
		{
			name: "multi local nodes",
			args: args{
				hosts:    []string{"node-0", "node-1", "node-2"},
				dataPath: "/var/lib/zookeeper",
			},
			want: helm3.M{
				"local": helm3.M{
					"0": helm3.M{
						"host": "node-0",
						"path": "/var/lib/zookeeper",
					},
					"1": helm3.M{
						"host": "node-1",
						"path": "/var/lib/zookeeper",
					},
					"2": helm3.M{
						"host": "node-2",
						"path": "/var/lib/zookeeper",
					},
				},
			},
		},
		{
			name: "multi hosted nodes",
			args: args{storageClassName: "standard"},
			want: helm3.M{"storageClassName": "standard"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HelmValuesStorageFor(tt.args.hosts, tt.args.dataPath, tt.args.storageClassName)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("HelmValuesStorageFor() got vs want: %v", d)
			}
		})
	}
}
