package proton_component

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"component-manage/pkg/models/types"
	"component-manage/pkg/models/types/components"
	store "taskrunner/pkg/store/proton"
	"taskrunner/pkg/store/proton/configuration"
	"taskrunner/test"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateRelease(t *testing.T) {
	t.SkipNow()
	tt := test.TestingT{T: t}
	kcli := fake.NewSimpleClientset()
	ss := trait.System{}
	s, err := NewServer(&store.ProtonClient{
		Namespace: "ut",
		ConfName:  "ut",
		Confkey:   "ut",
		Kcli:      kcli,
	}, ss, "")
	tt.AssertNil(err)
	ctx := context.Background()
	_, _ = kcli.CoreV1().Secrets("ut").Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ut",
			Namespace: "ut",
		},
	}, metav1.CreateOptions{})
	err = s.pcli.SetFullConf(ctx, &configuration.ClusterConfig{
		ApiVersion:          "v1",
		ResourceConnectInfo: &configuration.ResourceConnectInfo{},
	})
	tt.AssertNil(err)

	gin.SetMode(gin.ReleaseMode)

	node, rerr := os.Hostname()
	tt.AssertNil(rerr)

	testcase := []struct {
		Type string
		Rls  any
	}{
		{
			Type: "mariadb",
			Rls: types.ComponentMariaDB{
				Name: mariadbRlsName,
				Type: "mariadb",
				Params: &components.MariaDBComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/mariadb",
					Config: &struct {
						LowerCaseTableNames      *int   `json:"lower_case_table_names,omitempty" yaml:"lower_case_table_names,omitempty"`
						Thread_handling          string `json:"thread_handling,omitempty" yaml:"thread_handling,omitempty"`
						Innodb_buffer_pool_size  string `json:"innodb_buffer_pool_size" yaml:"innodb_buffer_pool_size"`
						Resource_requests_memory string `json:"resource_requests_memory" yaml:"resource_requests_memory"`
						Resource_limits_memory   string `json:"resource_limits_memory" yaml:"resource_limits_memory"`
					}{
						Innodb_buffer_pool_size:  "1G",
						Resource_requests_memory: "3G",
						Resource_limits_memory:   "3G",
					},
					Admin_user:      "root",
					Admin_passwd:    "FAKE_PASSWORD",
					AdminSecretName: "proton-mariadb-proton-rds",
					Username:        "anyshare",
					Password:        "FAKE_PASSWORD",
				},
			},
		},
		{
			Type: "etcd",
			Rls: types.ComponentETCD{
				Name: etcdRlsName,
				Type: "etcd",
				Params: &components.ETCDComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/etcd",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "policyengine",
			Rls: types.ComponentPolicyEngine{
				Name: pleRlsName,
				Type: "policyengine",
				Params: &components.PolicyEngineComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/policyengine",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
				Dependencies: &components.PolicyEngineComponentDependencies{
					ETCD: etcdRlsName,
				},
			},
		},

		{
			Type: "nebula",
			Rls: types.ComponentNebula{
				Name: nebulaRlsName,
				Type: "nebula",
				Params: &types.NebulaComponentParams{
					Hosts:           []string{node},
					Password:        "c12c0f2990187f2c3214a0e1",
					AdminSecretName: "nebula",
					DataPath:        "/sysvol/nebula",
				},
			},
		},
		{
			Type: "mongodb",
			Rls: types.ComponentMongoDB{
				Name: mongdbRlsName,
				Type: "mongodb",
				Params: &types.MongoDBComponentParams{
					ReplicaCount:    1,
					Hosts:           []string{node},
					Admin_user:      "root",
					Admin_passwd:    "FAKE_PASSWORD",
					Username:        "anyshare",
					Password:        "FAKE_PASSWORD",
					AdminSecretName: "mongodb-secret",
					Data_path:       "/sysvol/mongodb/mongodb_data",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.1",
							Memory: "128",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "redis",
			Rls: types.ComponentRedis{
				Name: redisRlsName,
				Type: "redis",
				Params: &components.RedisComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/redis",
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "opensearch",
			Rls: types.ComponentOpensearch{
				Name: opensearRlsName,
				Type: "opensearch",
				Params: &types.OpensearchComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/anyshare/opensearch",
					Mode:         "master",
					Config: components.OpensearchConfigs{
						JvmOptions:              "-Xmx0.2g -Xms0.2g",
						HanlpRemoteextDict:      "http://ecoconfig-private.anyshare:32128/api/ecoconfig/v1/word-list/remote_ext_dict",
						HanlpRemoteextStopwords: "http://ecoconfig-private.anyshare:32128/api/ecoconfig/v1/word-list/remote_ext_stopwords",
					},
					Settings: map[string]interface{}{
						"bootstrap.memory_lock":    "false",
						"action.auto_create_index": "-company*,-ar-*,+*",
					},
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					ExporterResources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.1",
							Memory: "200Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "zookeeper",
			Rls: types.ComponentZookeeper{
				Name: zkRlkName,
				Type: "zookeeper",
				Params: &types.ZookeeperComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					DataPath:     "/sysvol/sts/zk",
					Resources: components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "600Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					ExporterResources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.1",
							Memory: "200Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					StorageCapacity: "1Gi",
				},
			},
		},
		{
			Type: "kafka",
			Rls: types.ComponentKafka{
				Name: kafkaRlsName,
				Type: "kafka",
				Dependencies: &types.KafkaComponentDependencies{
					Zookeeper: zkRlkName,
				},
				Params: &types.KafkaComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					DataPath:     "/sysvol/sts/kafka",
					Resources: components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "200Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					ExporterResources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.1",
							Memory: "100Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					StorageCapacity: "1Gi",
				},
			},
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	s.RegistryHandler(r.Group(""))

	for _, ts := range testcase {
		{

			w := httptest.NewRecorder()
			bs, rerr := json.Marshal(ts.Rls)
			tt.AssertNil(rerr)
			req := httptest.NewRequest(http.MethodPut, "/components/release/"+ts.Type, bytes.NewReader(bs))
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr = io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

		{
			// do twice
			w := httptest.NewRecorder()
			bs, rerr := json.Marshal(ts.Rls)
			tt.AssertNil(rerr)
			req := httptest.NewRequest(http.MethodPut, "/components/release/"+ts.Type, bytes.NewReader(bs))
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr = io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

		{
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/components/release/"+ts.Type+"/"+ts.Type, nil)
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr := io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

	}
}

func TestInfoExternal(t *testing.T) {
	tt := test.TestingT{T: t}
	kcli := fake.NewSimpleClientset()
	ss := trait.System{}
	s, err := NewServer(&store.ProtonClient{
		Namespace: "ut",
		ConfName:  "ut",
		Confkey:   "ut",
		Kcli:      kcli,
	}, ss, "")
	tt.AssertNil(err)
	ctx := context.Background()
	_, _ = kcli.CoreV1().Secrets("ut").Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ut",
			Namespace: "ut",
		},
	}, metav1.CreateOptions{})
	err = s.pcli.SetFullConf(ctx, &configuration.ClusterConfig{
		ApiVersion:          "v1",
		ResourceConnectInfo: &configuration.ResourceConnectInfo{},
	})
	tt.AssertNil(err)

	gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	s.RegistryHandler(r.Group(""))

	// node, rerr := os.Hostname()
	// tt.AssertNil(rerr)

	testcase := []struct {
		Type string
		Rls  any
	}{
		{
			Type: "mq",
			Rls: MqFull{
				InfoMeta: InfoMeta{
					Name: "mq",
				},
				Info: configuration.MqInfo{
					SourceType: "external",
					MQType:     "kafka",
					MQHosts:    "test",
					MQPort:     123,
					Auth:       &components.KafkaAuth{},
				},
			},
		},
		{
			Type: "opensearch",
			Rls: OpensearchFull{
				InfoMeta: InfoMeta{
					Name: "opensearch",
				},
				OpensearchComponentInfo: components.OpensearchComponentInfo{
					SourceType: "external",
				},
			},
		},
		{
			Type: "mongodb",
			Rls: MongoDBFull{
				InfoMeta: InfoMeta{
					Name: "mongodb",
				},
				MongoDBComponentInfo: components.MongoDBComponentInfo{
					SourceType: "external",
				},
			},
		},
		{
			Type: "rds",
			Rls: RdsFull{
				InfoMeta: InfoMeta{
					Name: "rds",
				},
				MariaDBComponentInfo: components.MariaDBComponentInfo{
					SourceType: "external",
				},
			},
		},
		{
			Type: "redis",
			Rls: RedisFull{
				InfoMeta: InfoMeta{
					Name: "redis",
				},
				RedisComponentInfo: components.RedisComponentInfo{
					SourceType: "external",
				},
			},
		},
		{
			Type: "policyengine",
			Rls: PolicyEngineFull{
				InfoMeta: InfoMeta{
					Name: "policyengine",
				},
				PolicyEngineComponentInfo: components.PolicyEngineComponentInfo{
					SourceType: "external",
				},
			},
		},
		{
			Type: "etcd",
			Rls: EtcdFull{
				InfoMeta: InfoMeta{
					Name: "etcd",
				},
				ETCDComponentInfo: components.ETCDComponentInfo{
					SourceType: "external",
				},
			},
		},
	}
	for _, ts := range testcase {

		{
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/components/info/"+ts.Type+"/"+ts.Type, nil)
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 404 {
				bs, rerr := io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

		{

			w := httptest.NewRecorder()
			bs, rerr := json.Marshal(ts.Rls)
			tt.AssertNil(rerr)
			req := httptest.NewRequest(http.MethodPut, "/components/info/"+ts.Type, bytes.NewReader(bs))
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr = io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

		{
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/components/info/"+ts.Type+"/"+ts.Type, nil)
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr := io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

	}
}

func TestUpdateInfoRelease(t *testing.T) {
	t.SkipNow()
	tt := test.TestingT{T: t}
	kcli := fake.NewSimpleClientset()
	ss := trait.System{}
	s, err := NewServer(&store.ProtonClient{
		Namespace: "ut",
		ConfName:  "ut",
		Confkey:   "ut",
		Kcli:      kcli,
	}, ss, "")
	tt.AssertNil(err)
	ctx := context.Background()
	_, _ = kcli.CoreV1().Secrets("ut").Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ut",
			Namespace: "ut",
		},
	}, metav1.CreateOptions{})
	err = s.pcli.SetFullConf(ctx, &configuration.ClusterConfig{
		ApiVersion:          "v1",
		ResourceConnectInfo: &configuration.ResourceConnectInfo{},
	})
	tt.AssertNil(err)

	gin.SetMode(gin.ReleaseMode)

	node, rerr := os.Hostname()
	tt.AssertNil(rerr)

	testcase := []struct {
		Type string
		Rls  any
		Info MqFull
	}{
		{
			Type: "rds",
			Info: MqFull{
				InfoMeta: InfoMeta{
					Name: "rds",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "rds",
				},
			},
			Rls: types.ComponentMariaDB{
				Name: mariadbRlsName,
				Type: "mariadb",
				Params: &components.MariaDBComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/mariadb",
					Config: &struct {
						LowerCaseTableNames      *int   `json:"lower_case_table_names,omitempty" yaml:"lower_case_table_names,omitempty"`
						Thread_handling          string `json:"thread_handling,omitempty" yaml:"thread_handling,omitempty"`
						Innodb_buffer_pool_size  string `json:"innodb_buffer_pool_size" yaml:"innodb_buffer_pool_size"`
						Resource_requests_memory string `json:"resource_requests_memory" yaml:"resource_requests_memory"`
						Resource_limits_memory   string `json:"resource_limits_memory" yaml:"resource_limits_memory"`
					}{
						Innodb_buffer_pool_size:  "1G",
						Resource_requests_memory: "3G",
						Resource_limits_memory:   "3G",
					},
					Admin_user:      "root",
					Admin_passwd:    "FAKE_PASSWORD",
					AdminSecretName: "proton-mariadb-proton-rds",
					Username:        "anyshare",
					Password:        "FAKE_PASSWORD",
				},
			},
		},
		{
			Type: "etcd",
			Info: MqFull{
				InfoMeta: InfoMeta{
					Name: "etcd",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "etcd",
				},
			},
			Rls: types.ComponentETCD{
				Name: etcdRlsName,
				Type: "etcd",
				Params: &components.ETCDComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/etcd",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "policyengine",
			Info: MqFull{
				InfoMeta: InfoMeta{
					Name: "policyengine",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "policyengine",
				},
			},
			Rls: types.ComponentPolicyEngine{
				Name: pleRlsName,
				Type: "policyengine",
				Params: &components.PolicyEngineComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/policyengine",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
				Dependencies: &components.PolicyEngineComponentDependencies{
					ETCD: etcdRlsName,
				},
			},
		},
		{
			Type: "mongodb",
			Info: MqFull{
				InfoMeta: InfoMeta{
					Name: "mongodb",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "mongodb",
				},
			},
			Rls: types.ComponentMongoDB{
				Name: mongdbRlsName,
				Type: "mongodb",
				Params: &types.MongoDBComponentParams{
					ReplicaCount:    1,
					Hosts:           []string{node},
					Admin_user:      "root",
					Admin_passwd:    "FAKE_PASSWORD",
					Username:        "anyshare",
					Password:        "FAKE_PASSWORD",
					AdminSecretName: "mongodb-secret",
					Data_path:       "/sysvol/mongodb/mongodb_data",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.1",
							Memory: "128",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "redis",
			Info: MqFull{
				InfoMeta: InfoMeta{
					Name: "redis",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "reids",
				},
			},
			Rls: types.ComponentRedis{
				Name: redisRlsName,
				Type: "redis",
				Params: &components.RedisComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/sysvol/redis",
					Admin_user:   "root",
					Admin_passwd: "FAKE_PASSWORD",
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "opensearch",
			Info: MqFull{
				InfoMeta: InfoMeta{
					Name: "opensearch",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "opensearch",
				},
			},
			Rls: types.ComponentOpensearch{
				Name: opensearRlsName,
				Type: "opensearch",
				Params: &types.OpensearchComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					Data_path:    "/anyshare/opensearch",
					Mode:         "master",
					Config: components.OpensearchConfigs{
						JvmOptions:              "-Xmx0.2g -Xms0.2g",
						HanlpRemoteextDict:      "http://ecoconfig-private.anyshare:32128/api/ecoconfig/v1/word-list/remote_ext_dict",
						HanlpRemoteextStopwords: "http://ecoconfig-private.anyshare:32128/api/ecoconfig/v1/word-list/remote_ext_stopwords",
					},
					Settings: map[string]interface{}{
						"bootstrap.memory_lock":    "false",
						"action.auto_create_index": "-company*,-ar-*,+*",
					},
					Resources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "300Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					ExporterResources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.1",
							Memory: "200Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
				},
			},
		},
		{
			Type: "mq",
			Info: MqFull{
				InfoMeta: InfoMeta{
					Name: "mq",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "kafka",
				},
			},
			Rls: types.ComponentKafka{
				Name: kafkaRlsName,
				Type: "kafka",
				Dependencies: &types.KafkaComponentDependencies{
					Zookeeper: zkRlkName,
				},
				Params: &types.KafkaComponentParams{
					ReplicaCount: 1,
					Hosts:        []string{node},
					DataPath:     "/sysvol/sts/kafka",
					Resources: components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.2",
							Memory: "200Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					ExporterResources: &components.Resources{
						Limits: components.ResourceRequirements{
							CPU:    "0.1",
							Memory: "100Mi",
						},
						Requests: components.ResourceRequirements{
							CPU:    "0",
							Memory: "0",
						},
					},
					StorageCapacity: "1Gi",
				},
			},
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	s.RegistryHandler(r.Group(""))

	for _, ts := range testcase {
		{
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/components/info/"+ts.Type+"/"+ts.Type, nil)
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 404 {
				bs, rerr := io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}
		{

			w := httptest.NewRecorder()
			bs, rerr := json.Marshal(ts.Rls)
			tt.AssertNil(rerr)
			ts.Info.Instance = bs
			bs, rerr = json.Marshal(ts.Info)
			tt.AssertNil(rerr)
			req := httptest.NewRequest(http.MethodPut, "/components/info/"+ts.Type, bytes.NewReader(bs))
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr = io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

		{
			// test not set instance
			w := httptest.NewRecorder()
			ts.Info.Instance = nil
			bs, rerr := json.Marshal(ts.Info)
			tt.AssertNil(rerr)
			req := httptest.NewRequest(http.MethodPut, "/components/info/"+ts.Type, bytes.NewReader(bs))
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr = io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

		{
			// test cross release api
			w := httptest.NewRecorder()
			bs, rerr := json.Marshal(ts.Rls)
			tt.AssertNil(rerr)
			releseType := ts.Type
			switch ts.Type {
			case "rds":
				releseType = "mariadb"
			case "mq":
				releseType = "kafka"
			}
			req := httptest.NewRequest(http.MethodPut, "/components/release/"+releseType, bytes.NewReader(bs))
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr = io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

		{
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/components/info/"+ts.Type+"/"+ts.Type, nil)
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr := io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

	}

	{

		// test not set instance
		mq := MqFull{
			InfoMeta: InfoMeta{
				Name: "mq",
			},
			Info: configuration.MqInfo{
				SourceType: "internal",
				MQType:     "kafka",
			},
		}
		w := httptest.NewRecorder()
		kafka := types.ComponentKafka{
			Name: kafkaRlsName,
			Type: "kafka",
			Dependencies: &types.KafkaComponentDependencies{
				Zookeeper: zkRlkName,
			},
			Params: &types.KafkaComponentParams{
				ReplicaCount: 1,
				Hosts:        []string{node},
				DataPath:     "/sysvol/sts/kafka",
				Resources: components.Resources{
					Limits: components.ResourceRequirements{
						CPU:    "0.2",
						Memory: "200Mi",
					},
					Requests: components.ResourceRequirements{
						CPU:    "0",
						Memory: "0",
					},
				},
				ExporterResources: &components.Resources{
					Limits: components.ResourceRequirements{
						CPU:    "0.1",
						Memory: "100Mi",
					},
					Requests: components.ResourceRequirements{
						CPU:    "0",
						Memory: "0",
					},
				},
				StorageCapacity: "1Gi",
			},
		}
		zk := types.ComponentZookeeper{
			Name: zkRlkName,
			Type: "zookeeper",
			Params: &types.ZookeeperComponentParams{
				ReplicaCount: 1,
				Hosts:        []string{node},
				DataPath:     "/sysvol/sts/zk",
				Resources: components.Resources{
					Limits: components.ResourceRequirements{
						CPU:    "0.2",
						Memory: "600Mi",
					},
					Requests: components.ResourceRequirements{
						CPU:    "0",
						Memory: "0",
					},
				},
				ExporterResources: &components.Resources{
					Limits: components.ResourceRequirements{
						CPU:    "0.1",
						Memory: "200Mi",
					},
					Requests: components.ResourceRequirements{
						CPU:    "0",
						Memory: "0",
					},
				},
				StorageCapacity: "1Gi",
			},
		}
		bs, rerr := json.Marshal(kafka)
		tt.AssertNil(rerr)
		mq.Instance = bs
		bs, rerr = json.Marshal(zk)
		tt.AssertNil(rerr)
		mq.ZK = bs
		bs, rerr = json.Marshal(mq)
		tt.AssertNil(rerr)
		req := httptest.NewRequest(http.MethodPut, "/components/info/mq", bytes.NewReader(bs))
		r.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != 200 {
			bs, rerr = io.ReadAll(resp.Body)
			tt.AssertNil(rerr)
			t.Errorf("%s", bs)
			t.FailNow()
		}

		{
			mq := MqFull{
				InfoMeta: InfoMeta{
					Name: "mq",
				},
				Info: configuration.MqInfo{
					SourceType: "internal",
					MQType:     "nsq",
				},
			}
			bs, rerr = json.Marshal(mq)
			tt.AssertNil(rerr)
			req := httptest.NewRequest(http.MethodPut, "/components/info/mq", bytes.NewReader(bs))
			r.ServeHTTP(w, req)
			resp := w.Result()
			if resp.StatusCode != 200 {
				bs, rerr = io.ReadAll(resp.Body)
				tt.AssertNil(rerr)
				t.Errorf("%s", bs)
				t.FailNow()
			}
		}

	}
}
