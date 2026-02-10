package completion

import (
	"os"
	"testing"

	"github.com/go-test/deep"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestCompletionCS(t *testing.T) {
	tests := []struct {
		name   string
		config api.Config
		nodes  []configuration.Node
		want   configuration.Cs
	}{
		{
			name: "Local IPv4",
			config: api.Config{
				Clusters: map[string]*api.Cluster{
					"kubernetes": {
						Server: "https://proton-cs.lb.aishu.cn:8443",
					},
				},
			},
			nodes: []configuration.Node{
				{
					Name: "node-0",
					IP4:  "10.4.15.71",
				},
			},
			want: configuration.Cs{
				Provisioner: configuration.KubernetesProvisionerLocal,
				IPFamilies: []v1.IPFamily{
					v1.IPv4Protocol,
				},
				Addons: configuration.DefaultCSAddons,
			},
		},
		{
			name: "Local IPv6",
			config: api.Config{
				Clusters: map[string]*api.Cluster{
					"kubernetes": {
						Server: "https://proton-cs.lb.aishu.cn:8443",
					},
				},
			},
			nodes: []configuration.Node{
				{
					Name: "node-0",
					IP6:  "fe80::250:56ff:fec1:2271",
				},
			},
			want: configuration.Cs{
				Provisioner: configuration.KubernetesProvisionerLocal,
				IPFamilies: []v1.IPFamily{
					v1.IPv6Protocol,
				},
				Addons: configuration.DefaultCSAddons,
			},
		},
		{
			name: "Local DualStack",
			config: api.Config{
				Clusters: map[string]*api.Cluster{
					"kubernetes": {
						Server: "https://proton-cs.lb.aishu.cn:8443",
					},
				},
			},
			nodes: []configuration.Node{
				{
					Name: "node-0",
					IP4:  "10.4.15.71",
					IP6:  "fe80::250:56ff:fec1:2271",
				},
			},
			want: configuration.Cs{
				Provisioner: configuration.KubernetesProvisionerLocal,
				IPFamilies: []v1.IPFamily{
					v1.IPv4Protocol,
					v1.IPv6Protocol,
				},
				Addons: configuration.DefaultCSAddons,
			},
		},
		{
			name: "External",
			config: api.Config{
				Clusters: map[string]*api.Cluster{
					"huawei-cce": {
						Server: "https://cce-default.huaweicloud.cn:6443",
					},
				},
			},
			want: configuration.Cs{
				Provisioner: configuration.KubernetesProvisionerExternal,
				Addons:      configuration.DefaultCSAddons,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "kubeconfig-*.yaml")
			if err != nil {
				t.Fatalf("create kubeconfig file: %v", err)
			}
			defer os.Remove(f.Name())

			var oldRecommendedHomeFile = clientcmd.RecommendedHomeFile
			clientcmd.RecommendedHomeFile = f.Name()
			defer func() {
				clientcmd.RecommendedHomeFile = oldRecommendedHomeFile
			}()

			if err := clientcmd.WriteToFile(tt.config, f.Name()); err != nil {
				t.Fatalf("write kubeconfig file: %v, path, %v, config: %v", err, f.Name(), tt.config)
			}

			var c = configuration.Cs{}
			CompletionCS(&c, tt.nodes)
			for _, diff := range deep.Equal(c, tt.want) {
				t.Error(diff)
			}
		})
	}
}
func TestCompletionCSAddons(t *testing.T) {
	tests := []struct {
		name   string
		addons []configuration.CSAddonName
		want   []configuration.CSAddonName
	}{
		{
			name:   "not defined",
			addons: nil,
			want:   configuration.DefaultCSAddons,
		},
		{
			name:   "empty",
			addons: []configuration.CSAddonName{},
			want:   []configuration.CSAddonName{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := configuration.Cs{
				Addons: tt.addons,
			}
			CompletionCS(&c, nil)
			for _, diff := range deep.Equal(c.Addons, tt.want) {
				t.Errorf(".cs.addons, got vs want: %v", diff)
			}
		})
	}
}

func TestIsExternalKubernetes(t *testing.T) {
	tests := []struct {
		name   string
		config api.Config
		want   bool
	}{
		{
			name: "local",
			config: api.Config{
				Clusters: map[string]*api.Cluster{
					"kubernetes": {
						Server: "https://proton-cs.lb.aishu.cn:8443",
					},
				},
			},
			want: false,
		},
		{
			name: "external",
			config: api.Config{
				Clusters: map[string]*api.Cluster{
					"huawei-cce": {
						Server: "https://cce-default.huaweicloud.cn:6443",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "kubeconfig-*.yaml")
			if err != nil {
				t.Fatalf("create kubeconfig file: %v", err)
			}
			defer os.Remove(f.Name())

			if err := clientcmd.WriteToFile(tt.config, f.Name()); err != nil {
				t.Fatalf("write kubeconfig file: %v, path, %v, config: %v", err, f.Name(), tt.config)
			}

			if got := IsExternalKubernetes(f.Name()); got != tt.want {
				t.Errorf("Cs.IsExternalKubernetes() = %v, want %v", got, tt.want)
			}
		})
	}
}
