package grafana

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/grafana/helm"
)

func Test_valuesFor(t *testing.T) {
	const (
		registry         = "registry.example.org"
		host             = "host-example"
		dataPath         = "/var/lib/grafana"
		storageClassName = "standard"
		port             = 12450
	)
	var prometheus = &corev1.Service{Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: port}}}}
	var (
		quantity50m   = resource.MustParse("50m")
		quantity200m  = resource.MustParse("200m")
		quantity16Mi  = resource.MustParse("16Mi")
		quantity128Mi = resource.MustParse("128Mi")

		resources = &corev1.ResourceRequirements{
			Limits:   corev1.ResourceList{corev1.ResourceCPU: quantity200m, corev1.ResourceMemory: quantity128Mi},
			Requests: corev1.ResourceList{corev1.ResourceCPU: quantity50m, corev1.ResourceMemory: quantity16Mi},
		}
	)
	type args struct {
		spec       *configuration.Grafana
		registry   string
		prometheus *corev1.Service
	}
	tests := []struct {
		name string
		args args
		want *helm.Values
	}{
		{
			name: "example",
			args: args{
				spec:       &configuration.Grafana{Hosts: []string{host}, DataPath: dataPath, StorageClassName: storageClassName, Resources: resources},
				registry:   registry,
				prometheus: prometheus,
			},
			want: &helm.Values{
				Namespace:    "resource",
				Image:        helm.ValuesImage{Registry: registry},
				ReplicaCount: Replicas,
				Service:      helm.ValuesService{EnableDualStack: global.EnableDualStack, Grafana: helm.ValuesGrafanaService{Type: corev1.ServiceTypeNodePort, NodePort: NodePort}},
				Config:       helm.ValuesConfig{DataSource: helm.ValuesDataSource{Prometheus: valuesPrometheusFor(prometheus, "resource")}},
				Storage:      valuesStorageFor([]string{host}, dataPath, storageClassName, ""),
				Resources:    resources,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesFor(tt.args.spec, tt.args.registry, "resource", tt.args.prometheus)
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("valuesFor(), %v", d)
			}
		})
	}
}

func Test_valuesPrometheusFor(t *testing.T) {
	const (
		name           = "prometheus-example"
		namespaceOther = "other"
		port0          = 12450
		port1          = 12451
	)
	type args struct {
		prometheus *corev1.Service
	}
	tests := []struct {
		name string
		args args
		want helm.ValuesPrometheus
	}{
		{
			name: "single port",
			args: args{prometheus: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "resource"}, Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: port0}}}}},
			want: helm.ValuesPrometheus{Enabled: true, Protocol: helm.ValuesProtocolHTTP, Host: name, Port: port0},
		},
		{
			name: "multi ports",
			args: args{prometheus: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "resource"}, Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: port0}, {Port: port1}}}}},
			want: helm.ValuesPrometheus{Enabled: true, Protocol: helm.ValuesProtocolHTTP, Host: name, Port: port0},
		},
		{
			name: "prometheus in other namespace",
			args: args{prometheus: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespaceOther}, Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: port0}}}}},
			want: helm.ValuesPrometheus{Enabled: true, Protocol: helm.ValuesProtocolHTTP, Host: name + "." + namespaceOther, Port: port0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesPrometheusFor(tt.args.prometheus, "resource")
			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("valuesPrometheusFor(), %v", d)
			}
		})
	}
}

func Test_valuesStorageFor(t *testing.T) {
	const (
		host0            = "host-0"
		host1            = "host-1"
		dataPath         = "/var/lib/grafana"
		storageClassName = "standard"
	)
	type args struct {
		hosts            []string
		dataPath         string
		storageClassName string
	}
	tests := []struct {
		name string
		args args
		want helm.ValuesStorage
	}{
		{
			name: "local",
			args: args{hosts: []string{host0, host1}, dataPath: dataPath},
			want: helm.ValuesStorage{Local: map[string]helm.ValuesLocal{"0": {Host: host0, Path: dataPath}, "1": {Host: host1, Path: dataPath}}},
		},
		{
			name: "storage class",
			args: args{storageClassName: storageClassName},
			want: helm.ValuesStorage{StorageClassName: storageClassName},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valuesStorageFor(tt.args.hosts, tt.args.dataPath, tt.args.storageClassName, ""); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("valuesStorageFor() = %v, want %v", got, tt.want)
			}
		})
	}
}
