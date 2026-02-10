package servicepackage

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"helm.sh/helm/v3/pkg/chart"
)

func TestServicePackage_Load(t *testing.T) {
	tests := []struct {
		path    string
		want    *ServicePackage
		wantErr bool
	}{
		{
			path: "service-package-0",
			want: &ServicePackage{
				basePath: "testdata/service-package-0",
				charts: []Chart{
					{
						Path: "charts/example-chart-12a7",
						Metadata: chart.Metadata{
							Name:        "example-chart-0",
							Version:     "0.1.0",
							Description: "An example Helm chart for testing",
							APIVersion:  "v2",
							AppVersion:  "1.16.0",
							Type:        "application",
						},
					},
					{
						Path: "charts/example-chart-b398",
						Metadata: chart.Metadata{
							Name:        "example-chart-1",
							Version:     "0.1.2",
							Description: "An example Helm chart for testing",
							APIVersion:  "v2",
							AppVersion:  "1.16.1",
							Type:        "application",
						},
					},
					{
						Path: "charts/example-chart-0100",
						Metadata: chart.Metadata{
							Name:        "example-chart-1",
							Version:     "0.1.1",
							Description: "An example Helm chart for testing",
							APIVersion:  "v2",
							AppVersion:  "1.16.1",
							Type:        "application",
						},
					},
				},
			},
		},
		{
			path:    "non-exist-service-package",
			want:    &ServicePackage{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			obj := new(ServicePackage)
			err := obj.Load(filepath.Join("testdata", tt.path))
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(obj, tt.want) {
				t.Errorf("Load() = %v, want %v", obj, tt.want)
			}
			for _, diff := range deep.Equal(obj.charts, tt.want.charts) {
				t.Errorf("Load() difference between got and want: %v", diff)
			}
		})
	}
}

func TestServicePackage_Charts(t *testing.T) {
	type fields struct {
		charts Charts
	}
	tests := []struct {
		name   string
		fields fields
		want   Charts
	}{
		{
			name: "example",
			fields: fields{
				charts: Charts{
					{
						Path: "charts/test-chart-0",
						Metadata: chart.Metadata{
							Name:    "test-chart",
							Version: "1.0.0",
						},
					},
				},
			},
			want: Charts{
				{
					Path: "charts/test-chart-0",
					Metadata: chart.Metadata{
						Name:    "test-chart",
						Version: "1.0.0",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ServicePackage{
				charts: tt.fields.charts,
			}
			if got := p.Charts(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ServicePackage.Charts() = %v, want %v", got, tt.want)
			}
		})
	}
}
