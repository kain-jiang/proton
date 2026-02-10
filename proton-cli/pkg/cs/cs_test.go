package cs

import (
	"testing"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1/fake"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestIsControlPlaneNodeChanged(t *testing.T) {
	type args struct {
		new []string
		old []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{name: "same", args: args{[]string{"a", "b"}, []string{"a", "b"}}, want: false},
		{name: "reverse", args: args{[]string{"a", "b"}, []string{"b", "a"}}, want: false},
		{name: "add", args: args{[]string{"a", "b"}, []string{"a", "b", "c"}}, want: true},
		{name: "replace", args: args{[]string{"a", "b"}, []string{"a", "c"}}, want: true},
		{name: "delete", args: args{[]string{"a", "b"}, []string{"a"}}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsControlPlaneChanged(tt.args.new, tt.args.old); got != tt.want {
				t.Errorf("IsControlPlaneNodeChanged() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResetWhenConfIsNil(t *testing.T) {
	tests := []struct {
		name    string
		conf    *configuration.ClusterConfig
		wantErr bool
	}{
		{
			name:    "cluster conf is nil",
			conf:    nil,
			wantErr: true,
		},
		{
			name: "cs conf is nil",
			conf: &configuration.ClusterConfig{
				Cs: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := Cs{
				ClusterConf: tt.conf,
			}
			if err := cs.Reset(); (err != nil) != tt.wantErr {
				t.Errorf("Cs.Reset() err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetAdminConfFromMasterNode(t *testing.T) {
	tests := []struct {
		name     string
		conf     *configuration.ClusterConfig
		allNodes []v1alpha1.Interface
		wantErr  bool
	}{
		{
			name: "normal case",
			conf: &configuration.ClusterConfig{
				Cs: &configuration.Cs{
					Master: []string{"node-1"},
				},
			},
			allNodes: []v1alpha1.Interface{
				fake.NewForTesting(t, "node-1", []string{"/etc/kubernetes"}, []string{"/etc/kubernetes/admin.conf"}),
				fake.NewForTesting(t, "node-2", []string{}, []string{}),
			},
			wantErr: false,
		},
		{
			name: "source dir does not exist",
			conf: &configuration.ClusterConfig{
				Cs: &configuration.Cs{
					Master: []string{"node-1"},
				},
			},
			allNodes: []v1alpha1.Interface{
				fake.NewForTesting(t, "node-1", []string{}, []string{}),
				fake.NewForTesting(t, "node-2", []string{}, []string{}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("unimplemented")
		})
	}
}

func TestCopyAdminConfToWorkerNode(t *testing.T) {
	tests := []struct {
		name     string
		conf     *configuration.ClusterConfig
		allNodes []v1alpha1.Interface
		wantErr  bool
	}{
		{
			name: "normal case",
			conf: &configuration.ClusterConfig{
				Cs: &configuration.Cs{
					Master: []string{"node-1"},
				},
			},
			allNodes: []v1alpha1.Interface{
				fake.NewForTesting(t, "node-1", []string{}, []string{}),
				fake.NewForTesting(t, "node-2", []string{"/root/.kube"}, []string{}),
			},
			wantErr: false,
		},
		{
			name: "target dir does not exist",
			conf: &configuration.ClusterConfig{
				Cs: &configuration.Cs{
					Master: []string{"node-1"},
				},
			},
			allNodes: []v1alpha1.Interface{
				fake.NewForTesting(t, "node-1", []string{}, []string{}),
				fake.NewForTesting(t, "node-2", []string{}, []string{}),
			},
			wantErr: true,
		},
		{
			name: "old kube config exists",
			conf: &configuration.ClusterConfig{
				Cs: &configuration.Cs{
					Master: []string{"node-1"},
				},
			},
			allNodes: []v1alpha1.Interface{
				fake.NewForTesting(t, "node-1", []string{}, []string{}),
				fake.NewForTesting(t, "node-2", []string{"/root/.kube"}, []string{"/root/.kube/config"}),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("unimplemented")
		})
	}
}
