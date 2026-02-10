package configuration

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

func TestUploadToKubernetes(t *testing.T) {
	tests := []struct {
		name           string
		updateExisting bool
	}{
		{
			name: "create-new",
		},
		{
			name:           "update-existing",
			updateExisting: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := new(ClusterConfig)
			client := fake.NewSimpleClientset()
			if err := UploadToKubernetes(context.TODO(), cfg, client); err != nil {
				t.Fatalf("UploadToKubernetes() error = %v", err)
			}
			if tt.updateExisting {
				if err := UploadToKubernetes(context.TODO(), cfg, client); err != nil {
					t.Fatalf("UploadToKubernetes() error = %v", err)
				}
			}
		})
	}
}

// TODO: 完整的单元测试需要模拟 sftp 客户端及远程的文件系统
func TestRemoveOldProtonCLIDirIfExist(t *testing.T) {
	t.Skip("unimplemented")
}
