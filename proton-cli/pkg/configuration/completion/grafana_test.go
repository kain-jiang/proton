package completion

import (
	"testing"

	"github.com/go-test/deep"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/grafana"
)

func TestCompleteGrafana(t *testing.T) {
	tests := []struct {
		name       string
		spec, want *configuration.Grafana
	}{
		{
			name: "local and undefined data path",
			spec: &configuration.Grafana{Hosts: []string{"node-0"}},
			want: &configuration.Grafana{Hosts: []string{"node-0"}, DataPath: grafana.DefaultDataPath},
		},
		{
			name: "hosted and undefined data path",
			spec: &configuration.Grafana{},
			want: &configuration.Grafana{},
		},
		{
			name: "defined data path",
			spec: &configuration.Grafana{DataPath: "/var/lib/grafana"},
			want: &configuration.Grafana{DataPath: "/var/lib/grafana"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompleteGrafana(tt.spec)
			for _, d := range deep.Equal(tt.spec, tt.want) {
				t.Errorf("Grafana got != want: %v", d)
			}
		})
	}
}
