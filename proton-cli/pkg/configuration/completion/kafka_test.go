package completion

import (
	"testing"

	"github.com/go-test/deep"
	"helm.sh/helm/v3/pkg/chart"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

func TestCompleteKafka(t *testing.T) {
	tests := []struct {
		name   string
		obj    *configuration.Kafka
		charts servicepackage.Charts
		want   *configuration.Kafka
	}{
		{
			name: "unchanged",
			obj: &configuration.Kafka{
				Hosts: []string{
					"node-0",
					"node-1",
					"node-2",
				},
			},
			charts: servicepackage.Charts{
				{
					Path: "some-path",
					Metadata: chart.Metadata{
						Name:    "proton-kafka",
						Version: "1.2.3",
					},
				},
			},
			want: &configuration.Kafka{
				Hosts: []string{
					"node-0",
					"node-1",
					"node-2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompleteKafka(tt.obj, tt.charts)
			for _, diff := range deep.Equal(tt.obj, tt.want) {
				t.Errorf("TestCompleteKafka() difference of Kafka: %v", diff)
			}
		})
	}
}
