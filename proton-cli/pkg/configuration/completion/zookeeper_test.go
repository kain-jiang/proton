package completion

import (
	"testing"

	"github.com/go-test/deep"
	"helm.sh/helm/v3/pkg/chart"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

func TestCompleteZooKeeper(t *testing.T) {
	tests := []struct {
		name   string
		obj    *configuration.ZooKeeper
		charts servicepackage.Charts
		want   *configuration.ZooKeeper
	}{
		{
			name: "unchanged",
			obj: &configuration.ZooKeeper{
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
						Name:    "proton-zookeeper",
						Version: "1.2.3",
					},
				},
			},
			want: &configuration.ZooKeeper{
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
			CompleteZooKeeper(tt.obj, tt.charts)
			for _, diff := range deep.Equal(tt.obj, tt.want) {
				t.Errorf("TestCompleteZooKeeper() difference of ZooKeeper: %v", diff)
			}
		})
	}
}
