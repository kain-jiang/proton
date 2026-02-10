package keepalived

import (
	"testing"

	"github.com/go-test/deep"
)

func TestDeepCopy(t *testing.T) {
	tests := []struct {
		name   string
		values *HelmValues
	}{
		{
			name:   "nil",
			values: nil,
		},
		{
			name: "non nil",
			values: &HelmValues{
				Env: map[string]string{
					"language": "en_US.UTF-8",
					"timezone": "Asia/Shanghai",
				},
				Image: &HelmValuesImage{
					Registry: "registry.aishu.cn:15000",
				},
				Namespace: "resource",
				NodeSelector: map[string]string{
					"kubernetes.io/hostname": "vm-14-71-centos",
				},
				RBAC: &HelmValuesRBAC{
					Create: true,
				},
				VIP: &HelmValuesVIP{
					Interface:   "ens192",
					IP:          "10.10.14.71",
					ServiceName: "proton-mariadb-proton-rds-mariadb",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, diff := range deep.Equal(tt.values.DeepCopy(), tt.values) {
				t.Errorf("got vs orig: %v", diff)
			}
		})
	}
}
