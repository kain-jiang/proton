package cms

import (
	"reflect"
	"testing"

	"helm.sh/helm/v3/pkg/chart"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/universal"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/servicepackage"
)

func TestNewManager(t *testing.T) {
	type args struct {
		helm3          helm3.Client
		spec           *configuration.CMS
		registry       string
		servicePackage string
		charts         servicepackage.Charts
	}
	tests := []struct {
		name string
		args args
		want *universal.HelmV3Manager
	}{
		{
			name: "test",
			args: args{
				helm3:          nil,
				spec:           &configuration.CMS{},
				registry:       "test-registry",
				servicePackage: "/to/test/path",
				charts: servicepackage.Charts{
					servicepackage.Chart{
						Path: "chart.tgz",
						Metadata: chart.Metadata{
							Name:    ChartName,
							Version: "0.0.0",
						},
					},
				},
			},
			want: &universal.HelmV3Manager{
				Release:   ReleaseName,
				ChartFile: "/to/test/path/chart.tgz",
				Namespace: "resource",
				Helm3:     nil,
				Values: helm3.M{
					"image": helm3.M{
						"registry": "test-registry",
					},
					"serviceAccount": map[string]interface{}{
						"create": false,
						"name":   "serviceaccount",
					},
					"service": helm3.M{
						"protoncliNamespace": "proton",
					},
					"namespace":    "resource",
					"nodeSelector": (map[string]string)(nil),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewManager(tt.args.helm3, tt.args.spec, tt.args.registry, tt.args.servicePackage, tt.args.charts, "serviceaccount"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewManager() = %v, want %v", got, tt.want)
			}
		})
	}
}
