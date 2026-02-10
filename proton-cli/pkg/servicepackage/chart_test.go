package servicepackage

import (
	"reflect"
	"sort"
	"testing"

	"helm.sh/helm/v3/pkg/chart"
)

func TestCharts_Get(t *testing.T) {

	type args struct {
		name    string
		version string
	}
	tests := []struct {
		name   string
		charts Charts
		args   args
		want   *Chart
	}{
		{
			name: "unspecific-version",
			charts: Charts{
				{
					Path: "charts/test-chart-0",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "1.0.0",
					},
				},
				{
					Path: "charts/test-chart-1",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "2.0.0",
					},
				},
				{
					Path: "charts/test-chart-2",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "3.0.0",
					},
				},
			},
			args: args{
				name: "test-chart",
			},
			want: &Chart{
				Path: "charts/test-chart-2",
				Metadata: chart.Metadata{
					Name:    "test-chart",
					Version: "3.0.0",
				},
			},
		},
		{
			name: "unspecific-version-non-exist",
			charts: Charts{
				{
					Path: "charts/test-chart-0",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "1.0.0",
					},
				},
				{
					Path: "charts/test-chart-1",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "2.0.0",
					},
				},
				{
					Path: "charts/test-chart-2",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "3.0.0",
					},
				},
			},
			args: args{
				name: "test-chart-non-exist",
			},
		},
		{
			name: "specific-version",
			charts: Charts{
				{
					Path: "charts/test-chart-0",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "1.0.0",
					},
				},
				{
					Path: "charts/test-chart-1",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "2.0.0",
					},
				},
				{
					Path: "charts/test-chart-2",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "3.0.0",
					},
				},
			},
			args: args{
				name:    "test-chart",
				version: "2.0.0",
			},
			want: &Chart{
				Path: "charts/test-chart-1",
				Metadata: chart.Metadata{
					Name:    "test-chart",
					Version: "2.0.0",
				},
			},
		},
		{
			name: "specific-version-non-exist",
			charts: Charts{
				{
					Path: "charts/test-chart-0",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "1.0.0",
					},
				},
				{
					Path: "charts/test-chart-1",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "2.0.0",
					},
				},
				{
					Path: "charts/test-chart-2",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "3.0.0",
					},
				},
			},
			args: args{
				name:    "test-chart",
				version: "1.1.1",
			},
		},
	}
	for _, tt := range tests {
		sort.Sort(ByNameAndVersion(tt.charts))
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.charts.Get(tt.args.name, tt.args.version); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Charts.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
