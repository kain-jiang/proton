package validation

import (
	"testing"

	v1 "k8s.io/api/core/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateClusterConfig(t *testing.T) {
	type args struct {
		c *configuration.ClusterConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				c: &configuration.ClusterConfig{
					Deploy: &configuration.Deploy{
						Mode: "standard",
					},
					Firewall: configuration.Firewall{
						Mode: configuration.FirewallFirewalld,
					},
					Nodes: []configuration.Node{
						{
							Name: "node-0",
							IP4:  "192.168.0.1",
						},
						{
							Name: "node-1",
							IP4:  "192.168.0.2",
						},
						{
							Name: "node-2",
							IP4:  "192.168.0.3",
						},
					},
					Cs: &configuration.Cs{
						Provisioner: "local",
						IPFamilies: []v1.IPFamily{
							v1.IPv4Protocol,
						},
					},
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Proton_mariadb: &configuration.ProtonMariaDB{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Config: &configuration.ProtonMariaDBConfigs{
							Resource_requests_memory: "8G",
							Resource_limits_memory:   "9G",
						},
						Data_path: "/var/lib/mariadb",
					},
					Proton_mongodb: &configuration.ProtonDB{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/mongodb",
					},
					Proton_redis: &configuration.ProtonDB{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/redis",
					},
					Proton_mq_nsq: &configuration.ProtonDataConf{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/nsq",
					},
					Proton_policy_engine: &configuration.ProtonDataConf{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/policy-engine",
					},
					Proton_etcd: &configuration.ProtonDataConf{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/etcd",
					},
					OpenSearch: &configuration.OpenSearch{
						Mode: configuration.OpenSearchModeMaster,
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/opensearch",
						Settings: map[string]interface{}{
							"action.auto_create_index": "something",
							"bootstrap.memory_lock":    "something",
						},
					},
					Kafka: &configuration.Kafka{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
					ZooKeeper: &configuration.ZooKeeper{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
					Nebula: &configuration.Nebula{
						Hosts: []string{
							"node-0",
						},
						DataPath: "/var/lib/nebula",
						Password: "FAKE_PASSWORD",
					},
				},
			},
		},
		{
			name: "invalid-cr-missing",
			args: args{
				c: &configuration.ClusterConfig{
					Nodes: []configuration.Node{
						{
							Name: "node-0",
							IP4:  "192.168.0.1",
						},
						{
							Name: "node-1",
							IP4:  "192.168.0.2",
						},
						{
							Name: "node-2",
							IP4:  "192.168.0.3",
						},
					},
					Cs: &configuration.Cs{
						Provisioner: "local",
					},
					OpenSearch: &configuration.OpenSearch{
						Mode: configuration.OpenSearchModeMaster,
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/opensearch",
					},
					Kafka: &configuration.Kafka{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
					ZooKeeper: &configuration.ZooKeeper{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-opensearch",
			args: args{
				c: &configuration.ClusterConfig{
					Nodes: []configuration.Node{
						{
							Name: "node-0",
							IP4:  "192.168.0.1",
						},
						{
							Name: "node-1",
							IP4:  "192.168.0.2",
						},
						{
							Name: "node-2",
							IP4:  "192.168.0.3",
						},
					},
					Cs: &configuration.Cs{
						Provisioner: "local",
					},
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					OpenSearch: &configuration.OpenSearch{
						Mode: configuration.OpenSearchModeMaster,
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/opensearch",
					},
					Kafka: &configuration.Kafka{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
					ZooKeeper: &configuration.ZooKeeper{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-kafka",
			args: args{
				c: &configuration.ClusterConfig{
					Nodes: []configuration.Node{
						{
							Name: "node-0",
							IP4:  "192.168.0.1",
						},
						{
							Name: "node-1",
							IP4:  "192.168.0.2",
						},
						{
							Name: "node-2",
							IP4:  "192.168.0.3",
						},
					},
					Cs: &configuration.Cs{
						Provisioner: "local",
					},
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					OpenSearch: &configuration.OpenSearch{
						Mode: configuration.OpenSearchModeMaster,
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/opensearch",
					},
					Kafka: &configuration.Kafka{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
					ZooKeeper: &configuration.ZooKeeper{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-kafka-without-zookeeper",
			args: args{
				c: &configuration.ClusterConfig{
					Nodes: []configuration.Node{
						{
							Name: "node-0",
							IP4:  "192.168.0.1",
						},
						{
							Name: "node-1",
							IP4:  "192.168.0.2",
						},
						{
							Name: "node-2",
							IP4:  "192.168.0.3",
						},
					},
					Cs: &configuration.Cs{
						Provisioner: "local",
					},
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					OpenSearch: &configuration.OpenSearch{
						Mode: configuration.OpenSearchModeMaster,
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/opensearch",
					},
					Kafka: &configuration.Kafka{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-zookeeper",
			args: args{
				c: &configuration.ClusterConfig{
					Nodes: []configuration.Node{
						{
							Name: "node-0",
							IP4:  "192.168.0.1",
						},
						{
							Name: "node-1",
							IP4:  "192.168.0.2",
						},
						{
							Name: "node-2",
							IP4:  "192.168.0.3",
						},
					},
					Cs: &configuration.Cs{
						Provisioner: "local",
					},
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					OpenSearch: &configuration.OpenSearch{
						Mode: configuration.OpenSearchModeMaster,
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/var/lib/opensearch",
					},
					Kafka: &configuration.Kafka{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
					ZooKeeper: &configuration.ZooKeeper{
						Hosts: []string{
							"node-0",
							"node-1",
							"node-2",
						},
						Data_path: "/data/path",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateClusterConfig(tt.args.c); (errList != nil) != tt.wantErr {
				for i, err := range errList {
					t.Errorf("ValidateClusterConfig() errList[%d] = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateClusterConfigUpdate(t *testing.T) {
	type args struct {
		o *configuration.ClusterConfig
		n *configuration.ClusterConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid-none",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:        &configuration.Cs{},
					Kafka:     &configuration.Kafka{},
					ZooKeeper: &configuration.ZooKeeper{},
					Nebula:    &configuration.Nebula{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:        &configuration.Cs{},
					Kafka:     &configuration.Kafka{},
					ZooKeeper: &configuration.ZooKeeper{},
					Nebula:    &configuration.Nebula{},
				},
			},
		},
		{
			name: "invalid-uninstall-mariadb",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:             &configuration.Cs{},
					Proton_mariadb: &configuration.ProtonMariaDB{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-kafka",
			args: args{
				o: &configuration.ClusterConfig{
					Cs: &configuration.Cs{},
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Kafka:     &configuration.Kafka{},
					ZooKeeper: &configuration.ZooKeeper{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:        &configuration.Cs{},
					ZooKeeper: &configuration.ZooKeeper{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-zookeeper",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:        &configuration.Cs{},
					Kafka:     &configuration.Kafka{},
					ZooKeeper: &configuration.ZooKeeper{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:    &configuration.Cs{},
					Kafka: &configuration.Kafka{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-proton-redis",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:           &configuration.Cs{},
					Proton_redis: &configuration.ProtonDB{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-policy-engine",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:                   &configuration.Cs{},
					Proton_policy_engine: &configuration.ProtonDataConf{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-mq-nsq",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:            &configuration.Cs{},
					Proton_mq_nsq: &configuration.ProtonDataConf{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-opensearch",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:         &configuration.Cs{},
					OpenSearch: &configuration.OpenSearch{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-mongodb",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:             &configuration.Cs{},
					Proton_mongodb: &configuration.ProtonDB{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-etcd",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:          &configuration.Cs{},
					Proton_etcd: &configuration.ProtonDataConf{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-prometheus",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:         &configuration.Cs{},
					Prometheus: &configuration.Prometheus{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-grafana",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:      &configuration.Cs{},
					Grafana: &configuration.Grafana{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-uninstall-nebula",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:     &configuration.Cs{},
					Nebula: &configuration.Nebula{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs: &configuration.Cs{},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid-change-storagecapacity",
			args: args{
				o: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:             &configuration.Cs{},
					Proton_mariadb: &configuration.ProtonMariaDB{},
				},
				n: &configuration.ClusterConfig{
					Cr: &configuration.Cr{
						Local: &configuration.LocalCR{},
					},
					Cs:             &configuration.Cs{},
					Proton_mariadb: &configuration.ProtonMariaDB{StorageCapacity: "10Gi"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateClusterConfigUpdate(tt.args.o, tt.args.n); len(errList) > 1 || (errList != nil) != tt.wantErr {
				for i, err := range errList {
					t.Errorf("ValidateClusterConfigUpdate() errList[%d] = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}
