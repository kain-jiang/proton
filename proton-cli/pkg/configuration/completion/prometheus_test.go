package completion

import (
	"testing"

	"github.com/go-test/deep"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus"
)

func TestCompletePrometheus(t *testing.T) {
	tests := []struct {
		name       string
		spec, want *configuration.Prometheus
	}{
		{
			name: "local and undefined data path",
			spec: &configuration.Prometheus{Hosts: []string{"node-0"}},
			want: &configuration.Prometheus{Hosts: []string{"node-0"}, DataPath: prometheus.DefaultDataPath},
		},
		{
			name: "hosted and undefined data path",
			spec: &configuration.Prometheus{},
			want: &configuration.Prometheus{},
		},
		{
			name: "defined data path",
			spec: &configuration.Prometheus{DataPath: "/var/lib/prometheus"},
			want: &configuration.Prometheus{DataPath: "/var/lib/prometheus"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompletePrometheus(tt.spec)
			for _, d := range deep.Equal(tt.spec, tt.want) {
				t.Errorf("Prometheus got != want: %v", d)
			}
		})
	}
}
