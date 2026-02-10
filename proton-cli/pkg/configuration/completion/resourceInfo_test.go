package completion

import (
	"context"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/agiledragon/gomonkey"
	"github.com/go-test/deep"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	fake "k8s.io/client-go/kubernetes/typed/core/v1/fake"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/kafka"
)

func TestCompleteOldClusterConfFromSecret(t *testing.T) {
	type getInfo struct {
		data map[string]interface{}
		err  error
	}
	type want struct {
		result  *configuration.ClusterConfig
		wanterr error
	}
	tests := []struct {
		name     string
		dataList []getInfo
		param    *configuration.ClusterConfig
		want
	}{
		{
			name: "all-info-ok",

			dataList: []getInfo{
				{data: map[string]interface{}{
					"dbType":   "MariaDB",
					"host":     "mariadb-mariadb-master.resource",
					"hostRead": "mariadb-mariadb-cluster.resource",
					"password": "FAKE_PASSWORD",
					"port":     3330.0,
					"portRead": 3330.0,
					"type":     "Proton_MariaDB",
					"user":     "anyshare",
				}, err: nil},
				{data: map[string]interface{}{
					"sourceType": "test",
					"authSource": "anyshare",
					"host":       "mongodb-mongodb-0.mongodb-mongodb.resource",
					"options":    "",
					"password":   "FAKE_PASSWORD",
					"port":       28000.0,
					"replicaSet": "rs0",
					"ssl":        false,
					"user":       "anyshare",
				}, err: nil},
				{data: nil, err: nil},
				{data: map[string]interface{}{}, err: nil},
				{data: map[string]interface{}{
					"auth":          nil,
					"connectorType": "nsq",
					"mqHost":        "proton-mq-nsq-nsqd.resource",
					"mqLookupdHost": "proton-mq-nsq-nsqlookupd.resource",
					"mqLookupdPort": 4161.0,
					"mqPort":        4151.0,
					"mqType":        "nsq",
				}, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
			},
			param: &configuration.ClusterConfig{},
			want: want{
				result: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							RdsType:    configuration.MariaDB,
							Hosts:      "mariadb-mariadb-master.resource",
							Port:       3330,
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							HostsRead:  "mariadb-mariadb-cluster.resource",
							PortRead:   3330,
						},
						Mongodb: &configuration.MongodbInfo{
							SourceType: configuration.Internal,
							Hosts:      "mongodb-mongodb-0.mongodb-mongodb.resource",
							Port:       28000,
							ReplicaSet: "rs0",
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							SSL:        false,
							AuthSource: "anyshare",
							Options:    "",
						},
						Redis:      nil,
						OpenSearch: &configuration.OpensearchInfo{},
						Mq: &configuration.MqInfo{
							SourceType:     configuration.Internal,
							MqHosts:        "proton-mq-nsq-nsqd.resource",
							MqPort:         4151,
							MqLookupdHosts: "proton-mq-nsq-nsqlookupd.resource",
							MqLookupdPort:  4161,
							MqType:         "nsq",
							Auth:           nil,
						},
						PolicyEngine: nil,
						Etcd:         nil,
					},
				},
				wanterr: nil,
			},
		},
		{
			name: "mq-kafka-ok",
			dataList: []getInfo{
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: map[string]interface{}{
					"auth": map[string]interface{}{
						"username":  "aishu",
						"password":  "FAKE_PASSWORD",
						"mechanism": "PALIN",
					},
					"connectorType": "kafka",
					"mqHost":        "test1",
					"mqPort":        8080.0,
					"mqType":        "kafka",
				}, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
			},
			param: &configuration.ClusterConfig{},
			want: want{
				result: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							SourceType: configuration.External,
							MqHosts:    "test1",
							MqPort:     8080,
							MqType:     "kafka",
							Auth: &configuration.Auth{
								Username:  "aishu",
								Password:  "FAKE_PASSWORD",
								Mechanism: "PALIN",
							},
						},
					},
				},
			},
		},
		{
			name: "mq-mongodb-ok",
			dataList: []getInfo{
				{data: nil, err: nil},
				{data: map[string]interface{}{
					"authSource": "anyshare",
					"host":       "mongodb-mongodb-0.mongodb-mongodb.resource",
					"options": map[string]interface{}{
						"loglevel":       "warn",
						"maxConnections": 20.0,
					},
					"password":   "FAKE_PASSWORD",
					"port":       28000.0,
					"replicaSet": "rs0",
					"ssl":        false,
					"user":       "anyshare",
				}, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
			},
			param: &configuration.ClusterConfig{},
			want: want{
				result: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mongodb: &configuration.MongodbInfo{
							SourceType: configuration.Internal,
							Hosts:      "mongodb-mongodb-0.mongodb-mongodb.resource",
							Port:       28000,
							ReplicaSet: "rs0",
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							SSL:        false,
							AuthSource: "anyshare",
							Options: map[string]interface{}{
								"loglevel":       "warn",
								"maxConnections": 20.0,
							},
						},
					},
				},
			},
		},
		{
			name: "mq-mongodb-ok",
			dataList: []getInfo{
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: nil, err: nil},
				{data: map[string]interface{}{
					"auth":          nil,
					"connectorType": "nsq",
					"mqHost":        "proton-mq-nsq-nsqd.resource.svc.cluster.local",
					"mqLookupdHost": "proton-mq-nsq-nsqlookupd.resource.svc.cluster.local",
					"mqLookupdPort": 4161.0,
					"mqPort":        4151.0,
					"mqType":        "nsq",
				}, err: nil},

				{data: nil, err: nil},
				{data: nil, err: nil},
			},
			param: &configuration.ClusterConfig{},
			want: want{
				result: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							SourceType:     configuration.Internal,
							MqHosts:        "proton-mq-nsq-nsqd.resource.svc.cluster.local",
							MqPort:         4151,
							MqLookupdHosts: "proton-mq-nsq-nsqlookupd.resource.svc.cluster.local",
							MqLookupdPort:  4161,
							MqType:         "nsq",
							Auth:           nil,
						},
					},
				},
			},
		},
	}
	// rds mongodb redis es mq policy-engine etcd
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gomonkey.ApplyFuncSeq(getInfoFromSecret, []gomonkey.OutputCell{
				{Values: gomonkey.Params{tt.dataList[0].data, tt.dataList[0].err}},
				{Values: gomonkey.Params{tt.dataList[1].data, tt.dataList[1].err}},
				{Values: gomonkey.Params{tt.dataList[2].data, tt.dataList[2].err}},
				{Values: gomonkey.Params{tt.dataList[3].data, tt.dataList[3].err}},
				{Values: gomonkey.Params{tt.dataList[4].data, tt.dataList[4].err}},
				{Values: gomonkey.Params{tt.dataList[5].data, tt.dataList[5].err}},
				{Values: gomonkey.Params{tt.dataList[6].data, tt.dataList[6].err}},
			}).Reset()

			if err := CompleteOldClusterConfFromSecret(tt.param, &kubernetes.Clientset{}); err != nil {
				t.Error(err)
				return
			}

			for _, d := range deep.Equal(tt.param, tt.want.result) {
				t.Errorf("resourcConectInfo got != want: %v", d)
			}
		})
	}
}

func TestCompleteInternalInfo(t *testing.T) {
	tests := []struct {
		name  string
		c     *configuration.ClusterConfig
		wantc *configuration.ClusterConfig
	}{
		{
			name: "commplete-all-internal",
			c: &configuration.ClusterConfig{
				Cs: &configuration.Cs{
					Provisioner: configuration.KubernetesProvisionerLocal,
				},
				Proton_mariadb: &configuration.ProtonMariaDB{
					Hosts: []string{"node1"},
				},
				Proton_mongodb: &configuration.ProtonDB{
					Hosts: []string{"node1"},
				},
				Proton_redis: &configuration.ProtonDB{
					Hosts:        []string{"node1"},
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
				},
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Proton_policy_engine: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Proton_etcd: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Rds: &configuration.RdsInfo{
						SourceType: configuration.Internal,
						Username:   "FAKE_USERNAME",
						Password:   "FAKE_PASSWORD",
					},
					Mongodb: &configuration.MongodbInfo{
						SourceType: configuration.Internal,
						Username:   "FAKE_USERNAME",
						Password:   "FAKE_PASSWORD",
					},
					Redis: &configuration.RedisInfo{},
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
					},
					PolicyEngine: &configuration.PolicyEngineInfo{},
					Etcd:         &configuration.EtcdInfo{},
				},
			},
			wantc: &configuration.ClusterConfig{
				Cs: &configuration.Cs{
					Provisioner: configuration.KubernetesProvisionerLocal,
				},
				Proton_mariadb: &configuration.ProtonMariaDB{
					Hosts: []string{"node1"},
				},
				Proton_mongodb: &configuration.ProtonDB{
					Hosts: []string{"node1"},
				},
				Proton_redis: &configuration.ProtonDB{
					Hosts:        []string{"node1"},
					Admin_user:   "FAKE_USERNAME",
					Admin_passwd: "FAKE_PASSWORD",
				},
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Proton_policy_engine: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Proton_etcd: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Rds: &configuration.RdsInfo{
						SourceType: configuration.Internal,
						RdsType:    configuration.MariaDB,
						Hosts:      "mariadb-mariadb-master.resource",
						Port:       3330,

						Username:  "FAKE_USERNAME",
						Password:  "FAKE_PASSWORD",
						HostsRead: "mariadb-mariadb-cluster.resource",
						PortRead:  3330,
					},
					Mongodb: &configuration.MongodbInfo{
						SourceType: configuration.Internal,
						Hosts:      "mongodb-mongodb-0.mongodb-mongodb.resource",
						Port:       28000,

						ReplicaSet: "rs0",
						Username:   "FAKE_USERNAME",
						Password:   "FAKE_PASSWORD",

						SSL:        false,
						AuthSource: "anyshare",
					},
					Redis: &configuration.RedisInfo{},
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
						MqType:     configuration.Nsq,

						MqHosts: "proton-mq-nsq-nsqd.resource",
						MqPort:  4151,

						MqLookupdHosts: "proton-mq-nsq-nsqlookupd.resource",
						MqLookupdPort:  4161,
					},
					PolicyEngine: &configuration.PolicyEngineInfo{},
					Etcd:         &configuration.EtcdInfo{},
				},
			},
		},
		{
			name: "commplete-rds-username-password-use-resourceInfo",
			c: &configuration.ClusterConfig{
				Proton_mariadb: &configuration.ProtonMariaDB{
					Hosts:        []string{"node1"},
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Rds: &configuration.RdsInfo{
						SourceType: configuration.Internal,
						Username:   "FAKE_USERNAME",
						Password:   "FAKE_PASSWORD",
					},
				},
			},
			wantc: &configuration.ClusterConfig{
				Proton_mariadb: &configuration.ProtonMariaDB{
					Hosts:        []string{"node1"},
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Rds: &configuration.RdsInfo{
						SourceType: configuration.Internal,
						RdsType:    configuration.MariaDB,
						Hosts:      "mariadb-mariadb-master.resource",
						Port:       3330,

						Username:  "FAKE_USERNAME",
						Password:  "FAKE_PASSWORD",
						HostsRead: "mariadb-mariadb-cluster.resource",
						PortRead:  3330,
					},
				},
			},
		},
		{
			name: "commplete-rds-username-password-nil-in-resourceInfo",
			c: &configuration.ClusterConfig{
				Proton_mariadb: &configuration.ProtonMariaDB{
					Hosts:        []string{"node1"},
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Rds: &configuration.RdsInfo{
						SourceType: configuration.Internal,
					},
				},
			},
			wantc: &configuration.ClusterConfig{
				Proton_mariadb: &configuration.ProtonMariaDB{
					Hosts:        []string{"node1"},
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Rds: &configuration.RdsInfo{
						SourceType: configuration.Internal,
						RdsType:    configuration.MariaDB,
						Hosts:      "mariadb-mariadb-master.resource",
						Port:       3330,

						HostsRead: "mariadb-mariadb-cluster.resource",
						PortRead:  3330,
					},
				},
			},
		},
		{
			name: "commplete-mongodb-username-password-use-nil-resourceInfo",
			c: &configuration.ClusterConfig{
				Proton_mongodb: &configuration.ProtonDB{
					Hosts:        []string{"node1"},
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
				},
			},
			wantc: &configuration.ClusterConfig{
				Proton_mongodb: &configuration.ProtonDB{
					Hosts:        []string{"node1"},
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{},
			},
		},
		{
			name: "commplete-mq--use-nil-resourceInfo-only-kafka",
			c: &configuration.ClusterConfig{
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
			},
			wantc: &configuration.ClusterConfig{
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
						MqType:     configuration.KafkaType,

						MqHosts: "kafka-headless.resource",
						MqPort:  9097,
						Auth: &configuration.Auth{
							Username:  kafka.KafkaDefaultSSLUser,
							Password:  kafka.KafkaDefaultSSLPassword,
							Mechanism: configuration.Plain,
						},
					},
				},
			},
		},
		{
			name: "commplete-mq--use-resourceInfo-type-kafka-both-nsq-kafka",
			c: &configuration.ClusterConfig{
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
						MqType:     configuration.KafkaType,
					},
				},
			},
			wantc: &configuration.ClusterConfig{
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
						MqType:     configuration.KafkaType,

						MqHosts: "kafka-headless.resource",
						MqPort:  9097,
						Auth: &configuration.Auth{
							Username:  kafka.KafkaDefaultSSLUser,
							Password:  kafka.KafkaDefaultSSLPassword,
							Mechanism: configuration.Plain,
						},
					},
				},
			},
		},
		{
			name: "commplete-mq--use-nil-resourceInfo-both-nsq-kafka",
			c: &configuration.ClusterConfig{
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
			},
			wantc: &configuration.ClusterConfig{
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
						MqType:     configuration.KafkaType,

						MqHosts: "kafka-headless.resource",
						MqPort:  9097,

						Auth: &configuration.Auth{
							Username:  kafka.KafkaDefaultSSLUser,
							Password:  kafka.KafkaDefaultSSLPassword,
							Mechanism: configuration.Plain,
						},
					},
				},
			},
		},
		{
			name: "commplete-mq-use-nil-resourceInfo-only-kafka",
			c: &configuration.ClusterConfig{
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
			},
			wantc: &configuration.ClusterConfig{
				Kafka: &configuration.Kafka{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
						MqType:     configuration.KafkaType,

						MqHosts: "kafka-headless.resource",
						MqPort:  9097,
						Auth: &configuration.Auth{
							Username:  kafka.KafkaDefaultSSLUser,
							Password:  kafka.KafkaDefaultSSLPassword,
							Mechanism: configuration.Plain,
						},
					},
				},
			},
		},
		{
			name: "commplete-mq-use-nsq-hosts-not-nil",
			c: &configuration.ClusterConfig{
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Mq: &configuration.MqInfo{
						MqHosts:        "proton-mq-nsqd.resource.svc.cluster.local",
						MqLookupdHosts: "proton-mq-nsqlookupd.resource.svc.cluster.local",
					},
				},
			},
			wantc: &configuration.ClusterConfig{
				Proton_mq_nsq: &configuration.ProtonDataConf{
					Hosts: []string{"node1"},
				},
				ResourceConnectInfo: &configuration.ResourceConnectInfo{
					Mq: &configuration.MqInfo{
						SourceType: configuration.Internal,
						MqType:     configuration.Nsq,

						MqHosts:        "proton-mq-nsq-nsqd.resource",
						MqPort:         4151,
						MqLookupdHosts: "proton-mq-nsq-nsqlookupd.resource",
						MqLookupdPort:  4161,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompleteInternalInfo(tt.c)

			for _, d := range deep.Equal(tt.c, tt.wantc) {
				t.Errorf("commpleteInternalInfo() got != want: %v", d)
			}
		})
	}
}

func Test_getMongodbHosts(t *testing.T) {
	tests := []struct {
		name      string
		args      int
		wantHosts string
	}{
		{
			name:      "valid-one",
			args:      1,
			wantHosts: "mongodb-mongodb-0.mongodb-mongodb.resource",
		},
		{
			name:      "valid-many",
			args:      3,
			wantHosts: "mongodb-mongodb-0.mongodb-mongodb.resource,mongodb-mongodb-1.mongodb-mongodb.resource,mongodb-mongodb-2.mongodb-mongodb.resource",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMongodbHosts(tt.args)

			for _, d := range deep.Equal(got, tt.wantHosts) {
				t.Errorf("getMongodbHosts() got != wantHosts: %v", d)
			}
		})
	}

}

func Test_getInfoFromSecret(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
		k    *kubernetes.Clientset
	}
	type k8sRetrun struct {
		secret *v1.Secret
		err    error
	}
	tests := []struct {
		name      string
		args      args
		k8sRetrun k8sRetrun
		wantData  map[string]interface{}
		wantErr   bool
	}{
		{
			name: "get-rds-ok",
			args: args{
				ctx:  context.Background(),
				name: "rds",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("admin_key: RkFLRV9VU0VSTkFNRTpGQUtFX1BBU1NXT1JE\ndbType: MariaDB\nhost: mariadb-mariadb-master.resource\nhostRead: mariadb-mariadb-cluster.resource\npassword: FAKE_PASSWORD\nport: 3330\nportRead: 3330\ntype: Proton_MariaDB\nuser: anyshare\n"),
					},
				},
			},
			wantData: map[string]interface{}{
				"admin_key": "RkFLRV9VU0VSTkFNRTpGQUtFX1BBU1NXT1JE",
				"dbType":    "MariaDB",
				"host":      "mariadb-mariadb-master.resource",
				"hostRead":  "mariadb-mariadb-cluster.resource",
				"password":  "FAKE_PASSWORD",
				"port":      3330.0,
				"portRead":  3330.0,
				"type":      "Proton_MariaDB",
				"user":      "anyshare",
			},
		},

		{
			name: "get-mongodb-ok",
			args: args{
				ctx:  context.Background(),
				name: "mongodb",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("authSource: anyshare\nhost: mongodb-mongodb-0.mongodb-mongodb.resource,mongodb-mongodb-1.mongodb-mongodb.resource,mongodb-mongodb-2.mongodb-mongodb.resource\noptions: \"\"\npassword: FAKE_PASSWORD\nport: 28000\nreplicaSet: rs0\nssl: false\nuser: anyshare\n"),
					},
				},
			},
			wantData: map[string]interface{}{
				"authSource": "anyshare",
				"host":       "mongodb-mongodb-0.mongodb-mongodb.resource,mongodb-mongodb-1.mongodb-mongodb.resource,mongodb-mongodb-2.mongodb-mongodb.resource",
				"options":    "",
				"password":   "FAKE_PASSWORD",
				"port":       28000.0,
				"replicaSet": "rs0",
				"ssl":        false,
				"user":       "anyshare",
			},
		},
		{
			name: "get-redis-ok",
			args: args{
				ctx:  context.Background(),
				name: "redis",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("caName: \"\"\ncertName: \"\"\nconnectInfo:\n  masterGroupName: mymaster\n  password: FAKE_PASSWORD\n  sentinelHost: proton-redis-proton-redis-sentinel.resource\n  sentinelPassword: FAKE_PASSWORD\n  sentinelPort: 26379\n  sentinelUsername: root\n  username: root\nconnectType: sentinel\nenableSSL: false\nkeyName: \"\"\nsecretName: \"\"\n"),
					},
				},
			},
			wantData: map[string]interface{}{
				"caName":     "",
				"certName":   "",
				"enableSSL":  false,
				"keyName":    "",
				"secretName": "",
				"connectInfo": map[string]interface{}{
					"masterGroupName":  "mymaster",
					"password":         "FAKE_PASSWORD",
					"sentinelHost":     "proton-redis-proton-redis-sentinel.resource",
					"sentinelPassword": "FAKE_PASSWORD",
					"sentinelPort":     26379.0,
					"sentinelUsername": "FAKE_USERNAME",
					"username":         "FAKE_USERNAME",
				},
				"connectType": "sentinel",
			},
		},
		{
			name: "get-es-ok",
			args: args{
				ctx:  context.Background(),
				name: "es",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("host: opensearch-master.resource\npassword: FAKE_PASSWORD\nport: 9200\nprotocol: http\nuser: admin\nversion: 7.10.0\n"),
					},
				},
			},
			wantData: map[string]interface{}{
				"host":     "opensearch-master.resource",
				"password": "FAKE_PASSWORD",
				"port":     9200.0,
				"protocol": "http",
				"user":     "FAKE_USERNAME",
				"version":  "7.10.0",
			},
		},
		{
			name: "get-mq-ok",
			args: args{
				ctx:  context.Background(),
				name: "mq",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("auth: {}\nconnectorType: nsq\nmqHost: proton-mq-nsq-nsqd.resource.svc.cluster.local\nmqLookupdHost: proton-mq-nsq-nsqlookupd.resource.svc.cluster.local\nmqLookupdPort: 4161\nmqPort: 4151\nmqType: nsq\n"),
					},
				},
			},
			wantData: map[string]interface{}{
				"auth":          map[string]interface{}{},
				"connectorType": "nsq",
				"mqHost":        "proton-mq-nsq-nsqd.resource.svc.cluster.local",
				"mqLookupdHost": "proton-mq-nsq-nsqlookupd.resource.svc.cluster.local",
				"mqLookupdPort": 4161.0,
				"mqPort":        4151.0,
				"mqType":        "nsq",
			},
		},
		{
			name: "get-proton-policy-engine-ok",
			args: args{
				ctx:  context.Background(),
				name: "proton-policy-engine",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("proton-policy-engine:\n  host: proton-policy-engine-proton-policy-engine-cluster.resource\n  port: 9800\n"),
					},
				},
			},
			wantData: map[string]interface{}{
				"proton-policy-engine": map[string]interface{}{
					"host": "proton-policy-engine-proton-policy-engine-cluster.resource",
					"port": 9800.0,
				},
			},
		},
		{
			name: "get-proton-etcd-ok",
			args: args{
				ctx:  context.Background(),
				name: "proton-etcd",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("proton-etcd:\n  host: proton-etcd.resource\n  port: 2379\n  secret: etcdssl-secret\n"),
					},
				},
			},
			wantData: map[string]interface{}{
				"proton-etcd": map[string]interface{}{
					"host":   "proton-etcd.resource",
					"port":   2379.0,
					"secret": "etcdssl-secret",
				},
			},
		},
		{
			name: "get-secret-no-default-yaml",
			args: args{
				ctx:  context.Background(),
				name: "proton-etcd",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"test.yaml": []byte("proton-etcd:\n  host: proton-etcd.resource\n  port: 2379\n  secret: etcdssl-secret\n"),
					},
				},
			},
			wantData: map[string]interface{}{},
		},
		{
			name: "get-secret-unmarshal-failed",
			args: args{
				ctx:  context.Background(),
				name: "proton-etcd",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				secret: &v1.Secret{
					Data: map[string][]byte{
						"default.yaml": []byte("proton-etcd:\nhost: proton-etcd.resource\n  port: 2379\n  secret: etcdssl-secret\n"),
					},
				},
			},
			wantData: map[string]interface{}{},
			wantErr:  true,
		},
		{
			name: "get-secret-not-find",
			args: args{
				ctx:  context.Background(),
				name: "proton-etcd",
				k:    &kubernetes.Clientset{},
			},
			k8sRetrun: k8sRetrun{
				err: errors.NewNotFound(schema.GroupResource{}, "not-found"),
			},
			wantData: nil,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer gomonkey.ApplyMethod(reflect.TypeOf(&kubernetes.Clientset{}), "CoreV1", func(_ *kubernetes.Clientset) core.CoreV1Interface {
				return &core.CoreV1Client{}
			}).ApplyMethod(reflect.TypeOf(&core.CoreV1Client{}), "Secrets", func(_ *core.CoreV1Client, _ string) core.SecretInterface {
				return &fake.FakeSecrets{}
			}).ApplyMethod(reflect.TypeOf(&fake.FakeSecrets{}), "Get", func(_ *fake.FakeSecrets, _ context.Context, _ string, _ metav1.GetOptions) (*v1.Secret, error) {
				return tt.k8sRetrun.secret, tt.k8sRetrun.err
			}).Reset()

			gotData, err := getInfoFromSecret(tt.args.ctx, tt.args.name, tt.args.k)
			if (err != nil) != tt.wantErr {
				t.Errorf("getInfoFromSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, d := range deep.Equal(gotData, tt.wantData) {
				t.Errorf("getInfoFromSecret() got != want: %v", d)
			}

		})
	}
}
