package helm

import (
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_resourcesFor(t *testing.T) {
	type args struct {
		resources *corev1.ResourceRequirements
	}
	tests := []struct {
		name string
		args args
		want *Resources
	}{
		{
			name: "defined",
			args: args{
				resources: &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
					Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
				},
			},
			want: &Resources{
				Store: &corev1.ResourceRequirements{
					Limits:   map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("2")},
					Requests: map[corev1.ResourceName]resource.Quantity{corev1.ResourceCPU: resource.MustParse("1")},
				},
			},
		},
		{
			name: "undefined",
			args: args{
				resources: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resourcesFor(tt.args.resources)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("resourcesFor() got != want: %v", d)
			}
		})
	}
}
