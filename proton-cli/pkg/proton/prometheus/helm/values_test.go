package helm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	"sigs.k8s.io/yaml"
)

func TestValues_ToMsap(t *testing.T) {
	const (
		namespace = "prometheus"

		registry = "registry.example.org"

		host0 = "node-0"
		host1 = "node-1"

		dataPath = "/var/lib/prometheus"

		storageClassName = "standard"
	)
	var (
		image = ValuesImage{Registry: registry}

		serviceSingleStack = ValuesService{EnableDualStack: false}
		serviceDualStack   = ValuesService{EnableDualStack: true}

		local0 = ValuesLocal{Host: host0, Path: dataPath}
		local1 = ValuesLocal{Host: host1, Path: dataPath}

		storageLocalSingle = ValuesStorage{Local: map[string]ValuesLocal{"0": local0}}
		storageLocalMulti  = ValuesStorage{Local: map[string]ValuesLocal{"0": local0, "1": local1}}
		storageHosted      = ValuesStorage{StorageClassName: storageClassName}

		ProtonETCDResultSecretName = "etcdssl-secret-for-prometheus"
		ProtonETCDResultCAName     = "ca-protonetcd.crt"
		ProtonETCDResultCertName   = "prometheus-metrics-protonetcd.crt"
		ProtonETCDResultKeyName    = "prometheus-metrics-protonetcd.key"
		K8SETCDResultSecretName    = "k8s-etcdssl-secret-for-prometheus"
		K8SETCDResultCAName        = "ca-k8setcd.crt"
		K8SETCDResultCertName      = "prometheus-metrics-k8setcd.crt"
		K8SETCDResultKeyName       = "prometheus-metrics-k8setcd.key"
		predefSecretStruct         = ValuesSecret{
			ProtonEtcd: ValuesEtcdCertInfo{
				Enabled:    true,
				SecretName: ProtonETCDResultSecretName,
				CaName:     ProtonETCDResultCAName,
				CertName:   ProtonETCDResultCertName,
				KeyName:    ProtonETCDResultKeyName,
			},
			K8sEtcd: ValuesEtcdCertInfo{
				Enabled:    true,
				SecretName: K8SETCDResultSecretName,
				CaName:     K8SETCDResultCAName,
				CertName:   K8SETCDResultCertName,
				KeyName:    K8SETCDResultKeyName,
			},
		}
	)
	type fields struct {
		ReplicaCount int
		Service      ValuesService
		Storage      ValuesStorage
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "local kubernetes single node",
			fields: fields{
				ReplicaCount: 1,
				Service:      serviceSingleStack,
				Storage:      storageLocalSingle,
			},
			want: "local-single-node.yaml",
		},
		{
			name: "local kubernetes multi nodes",
			fields: fields{
				ReplicaCount: 2,
				Service:      serviceSingleStack,
				Storage:      storageLocalMulti,
			},
			want: "local-multi-nodes.yaml",
		},
		{
			name: "hosted kubernetes",
			fields: fields{
				ReplicaCount: 2,
				Service:      serviceSingleStack,
				Storage:      storageHosted,
			},
			want: "hosted.yaml",
		},
		{
			name: "dual stack",
			fields: fields{
				ReplicaCount: 2,
				Service:      serviceDualStack,
				Storage:      storageLocalMulti,
			},
			want: "dula-stack.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var want map[string]interface{}
			if b, err := os.ReadFile(filepath.Join("testdata", tt.want)); err != nil {
				t.Fatal(err)
			} else if err := yaml.Unmarshal(b, &want); err != nil {
				t.Fatal(err)
			}

			v := &Values{
				Namespace:    namespace,
				Image:        image,
				ReplicaCount: tt.fields.ReplicaCount,
				Service:      tt.fields.Service,
				Storage:      tt.fields.Storage,
				Secret:       predefSecretStruct,
			}
			got := v.ToMap()

			for _, d := range deep.Equal(got, want) {
				t.Errorf("Values.ToMap() %v", d)
			}
		})
	}
}
