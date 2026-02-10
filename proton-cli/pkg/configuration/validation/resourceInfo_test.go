package validation

import (
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func Test_validateRds(t *testing.T) {
	type args struct {
		rds     *configuration.ClusterConfig
		fldPath *field.Path
	}
	const (
		invalidSourceType configuration.SourceType = "External"
		invalidRdsType    configuration.RDSType    = "Mariadb"
	)

	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "valid-external",
			args: args{
				rds: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.External,
							RdsType:    configuration.MariaDB,
							Hosts:      "test1,test2",
							Port:       3330,
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: nil,
		},
		{
			name: "valid-internal",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{
						Hosts: []string{"node1", "node2"},
					},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							RdsType:    configuration.MariaDB,
							Hosts:      "test1,test2",
							Port:       3330,
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: nil,
		},
		{
			name: "invalid-external",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{
						Hosts: []string{"node1", "node2"},
					},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.External,
							RdsType:    configuration.MariaDB,
							Hosts:      "test1,test2",
							Port:       3330,
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: nil,
		},
		{
			name: "invalid-rds-type",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							RdsType:    invalidRdsType,
							Hosts:      "test1",
							Port:       3330,
							Username:   "test",
							Password:   "test",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.rds.rds_type",
					BadValue: invalidRdsType,
					Detail:   "valid rdsType: MariaDB MySQL GoldenDB DM8 TiDB KDB9",
				},
			},
		},
		{
			name: "required-rds-type",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							Hosts:      "test1",
							Port:       3330,
							Username:   "test",
							Password:   "test",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.rds_type",
					BadValue: "",
				},
			},
		},
		{
			name: "required-hosts",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							RdsType:    configuration.MariaDB,
							Port:       3330,
							Username:   "test",
							Password:   "test",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.hosts",
					BadValue: "",
				},
			},
		},
		{
			name: "invalid-port",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							RdsType:    configuration.MariaDB,
							Hosts:      "test1",
							Port:       333000,
							Username:   "test",
							Password:   "test",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.rds.port",
					BadValue: 333000,
					Detail:   "valid port: 0-65535",
				},
			},
		},
		{
			name: "required-username",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							RdsType:    configuration.MariaDB,
							Hosts:      "test1",
							Port:       3330,
							Password:   "test",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.username",
					BadValue: "",
				},
			},
		},
		{
			name: "required-password",
			args: args{
				rds: &configuration.ClusterConfig{
					Proton_mariadb: &configuration.ProtonMariaDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{
							SourceType: configuration.Internal,
							RdsType:    configuration.MariaDB,
							Hosts:      "test1",
							Port:       3330,
							Username:   "test",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.password",
					BadValue: "",
				},
			},
		},
		{
			name: "required-rds",
			args: args{
				rds: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-all",
			args: args{
				rds: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Rds: &configuration.RdsInfo{},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("rds"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.username",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.password",
					BadValue: "",
				},

				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.rds.port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.rds.rds_type",
					BadValue: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := validateRds(tt.args.rds.ResourceConnectInfo.Rds, tt.args.rds, tt.args.fldPath)

			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("validateRds() allErrs != tt.wantAllErrs, %v", d)
			}
		})
	}
}

func Test_validateMongodb(t *testing.T) {
	type args struct {
		c       *configuration.ClusterConfig
		fldPath *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "valid-internal",
			args: args{
				c: &configuration.ClusterConfig{
					Proton_mongodb: &configuration.ProtonDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mongodb: &configuration.MongodbInfo{
							SourceType: configuration.Internal,
							Hosts:      "test1,test2",
							Port:       3330,
							ReplicaSet: "rs0",
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							SSL:        true,
							AuthSource: "anyshare",
							Options: map[string]string{
								"test-key": "test-value",
							},
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mongodb"),
			},
			wantAllErrs: nil,
		},
		{
			name: "valid-external",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mongodb: &configuration.MongodbInfo{
							SourceType: configuration.External,
							Hosts:      "test1,test2",
							Port:       3330,
							ReplicaSet: "rs0",
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							SSL:        true,
							AuthSource: "anyshare",
							Options: map[string]string{
								"test-key": "test-value",
							},
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mongodb"),
			},
			wantAllErrs: nil,
		},
		{
			name: "invalid-external",
			args: args{
				c: &configuration.ClusterConfig{
					Proton_mongodb: &configuration.ProtonDB{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mongodb: &configuration.MongodbInfo{
							SourceType: configuration.External,
							Hosts:      "test1,test2",
							Port:       3330,
							ReplicaSet: "rs0",
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							SSL:        true,
							AuthSource: "anyshare",
							Options: map[string]string{
								"test-key": "test-value",
							},
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mongodb"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-mongodb",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mongodb"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-all",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mongodb: &configuration.MongodbInfo{},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mongodb"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mongodb.username",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mongodb.password",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mongodb.hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mongodb.port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mongodb.replica_set",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mongodb.auth_source",
					BadValue: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := validateMongodb(tt.args.c.ResourceConnectInfo.Mongodb, tt.args.c, tt.args.fldPath)

			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("validateMongodb() gotAllErrs != wantAllErrs; %v", d)
			}

		})

	}
}

func Test_validateRedis(t *testing.T) {
	type args struct {
		c       *configuration.ClusterConfig
		fldPath *field.Path
	}
	const (
		invalidConnectType configuration.ConnectType = "master-slaver"
	)
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "valid-sentinel",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{
							SourceType:       configuration.External,
							ConnectType:      configuration.SentinelMode,
							Username:         "FAKE_USERNAME",
							Password:         "FAKE_PASSWORD",
							SentinelHosts:    "sentinel.resource",
							SentinelPort:     2379,
							SentinelUsername: "FAKE_USERNAME",
							SentinelPassword: "FAKE_PASSWORD",
							MasterGroupName:  "mymaster",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: nil,
		},
		{
			name: "invalid-connect-type",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{
							SourceType:       configuration.External,
							ConnectType:      invalidConnectType,
							Username:         "FAKE_USERNAME",
							Password:         "FAKE_PASSWORD",
							SentinelHosts:    "sentinel.resource",
							SentinelPort:     2379,
							SentinelUsername: "FAKE_USERNAME",
							SentinelPassword: "FAKE_PASSWORD",
							MasterGroupName:  "mymaster",
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.redis.connect_type",
					BadValue: invalidConnectType,
					Detail:   "valid connectType: sentinel master-slave standalone",
				},
			},
		},
		{
			name: "required-sentinel",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{
							ConnectType: configuration.SentinelMode,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.redis.sentinel_hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.redis.sentinel_port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.redis.master_group_name",
					BadValue: "",
				},
			},
		},
		{
			name: "valid-master-slave",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{
							SourceType:  configuration.External,
							ConnectType: configuration.MasterSlaverMode,
							Username:    "FAKE_USERNAME",
							Password:    "FAKE_PASSWORD",
							MasterHosts: "master.resource",
							MasterPort:  26379,
							SlaveHosts:  "slave-resource",
							SlavePort:   26380,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-master-slave",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{
							ConnectType: configuration.MasterSlaverMode,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.redis.master_hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.redis.master_port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.redis.slave_hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.redis.slave_port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
			},
		},
		{
			name: "valid-standAlone",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{
							SourceType:  configuration.External,
							ConnectType: configuration.StandAlonelMode,
							Username:    "FAKE_USERNAME",
							Password:    "FAKE_PASSWORD",
							Hosts:       "redis.resource",
							Port:        26379,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-standalone",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{
							ConnectType: configuration.StandAlonelMode,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.redis.hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.redis.port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
			},
		},
		{
			name: "invalid-all-nil",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Redis: &configuration.RedisInfo{},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("redis"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.redis.connect_type",
					BadValue: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotAllErrs := validateRedis(tt.args.c.ResourceConnectInfo.Redis, tt.args.c, tt.args.fldPath)

			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("validateRedis() gotAllErrs != tt.wantAllErrs, %v", d)
			}
		})
	}
}

func Test_validateMq(t *testing.T) {
	type args struct {
		c       *configuration.ClusterConfig
		fldPath *field.Path
	}
	const (
		invalidSourceType configuration.SourceType = "Internal"
		invalidMqType     configuration.MqType     = "htp2.0"
		invalidMechanism  configuration.MechanType = "plain"
	)
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "kafka-valid",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							SourceType: configuration.External,
							MqType:     configuration.KafkaType,
							MqHosts:    "test1",
							MqPort:     8080,
							Auth: &configuration.Auth{
								Username:  "FAKE_USERNAME",
								Password:  "FAKE_PASSWORD",
								Mechanism: configuration.Plain,
							},
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mq"),
			},
			wantAllErrs: nil,
		},
		{
			name: "kafka-valid-unassigned-auth",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							SourceType: configuration.External,
							MqType:     configuration.KafkaType,
							MqHosts:    "test1",
							MqPort:     8080,
							Auth:       &configuration.Auth{},
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mq"),
			},
			wantAllErrs: nil,
		},
		{
			name: "all-invalid",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							SourceType:     invalidSourceType,
							MqType:         invalidMqType,
							MqHosts:        "test1",
							MqPort:         808000,
							MqLookupdHosts: "test1",
							MqLookupdPort:  8080,
							Auth: &configuration.Auth{
								Username:  "FAKE_USERNAME",
								Password:  "FAKE_PASSWORD",
								Mechanism: invalidMechanism,
							},
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mq"),
			},

			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.source_type",
					BadValue: invalidSourceType,
					Detail:   "valid source_type: internal external",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.mq_port",
					BadValue: 808000,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.mq_type",
					BadValue: invalidMqType,
					Detail:   "valid mqType: nsq kafka Tonglink htp20 bmq htp202",
				},
				{
					Type:     field.ErrorTypeForbidden,
					Field:    "resourceConnectInfo.mq.auth",
					BadValue: "",
					Detail:   "auth only used by kafka;value: &{aishu FAKE_PASSWORD plain}",
				},
			},
		},
		{
			name: "kafka-invalid",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							SourceType:     invalidSourceType,
							MqType:         configuration.KafkaType,
							MqHosts:        "test1",
							MqPort:         808000,
							MqLookupdHosts: "test1",
							MqLookupdPort:  8080,
							Auth: &configuration.Auth{
								Username:  "FAKE_USERNAME",
								Password:  "FAKE_PASSWORD",
								Mechanism: invalidMechanism,
							},
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mq"),
			},

			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.source_type",
					BadValue: invalidSourceType,
					Detail:   "valid source_type: internal external",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.mq_port",
					BadValue: 808000,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.auth.mechanism",
					BadValue: invalidMechanism,
					Detail:   "valid mq MechanType: PLAIN SCRAM-SHA-512 SCRAM-SHA-256",
				},
			},
		},
		{
			name: "kafka-all-nil",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							MqType: configuration.KafkaType,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mq"),
			},

			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mq.source_type",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mq.mq_hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.mq_port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
			},
		},
		{
			name: "nsq-all-nil",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Mq: &configuration.MqInfo{
							MqType: configuration.Nsq,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("mq"),
			},

			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mq.source_type",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mq.mq_hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.mq_port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.mq.mq_lookupd_hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.mq.mq_lookupd_port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := validateMq(tt.args.c.ResourceConnectInfo.Mq, tt.args.c, tt.args.fldPath)

			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("validateMq() gotAllErrs != tt.wantAllErrs, %v", d)
			}
		})
	}
}

func Test_validateSearchEngine(t *testing.T) {
	type args struct {
		c       *configuration.ClusterConfig
		fldPath *field.Path
	}
	const invalidVersion configuration.Version = "1.1.1"
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "valid",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						OpenSearch: &configuration.OpensearchInfo{
							SourceType: configuration.External,
							Hosts:      "opensearch.resource",
							Port:       9800,
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							Protocol:   "http",
							Version:    configuration.Version564,
						},
					},
				},

				fldPath: field.NewPath("resourceConnectInfo").Child("search_engine"),
			},
			wantAllErrs: nil,
		},
		{
			name: "invalid-version",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						OpenSearch: &configuration.OpensearchInfo{
							SourceType: configuration.External,
							Hosts:      "opensearch.resource",
							Port:       9800,
							Username:   "FAKE_USERNAME",
							Password:   "FAKE_PASSWORD",
							Protocol:   "http",
							Version:    invalidVersion,
						},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("search_engine"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.search_engine.version",
					BadValue: invalidVersion,
					Detail:   "supported version: 5.6.4 7.10.0",
				},
			},
		},
		{
			name: "required-all",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						OpenSearch: &configuration.OpensearchInfo{},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("search_engine"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.search_engine.hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.search_engine.port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.search_engine.username",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.search_engine.password",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.search_engine.version",
					BadValue: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := validateSearchEngine(tt.args.c.ResourceConnectInfo.OpenSearch, tt.args.c, tt.args.fldPath)

			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("validateSearchEngine() gotAllErrs != tt.wantAllErrs, %v", d)
			}
		})
	}
}

func Test_validatePolicyEngine(t *testing.T) {
	type args struct {
		c       *configuration.ClusterConfig
		fldPath *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "valid-internal",
			args: args{
				c: &configuration.ClusterConfig{
					Proton_policy_engine: &configuration.ProtonDataConf{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						PolicyEngine: &configuration.PolicyEngineInfo{
							SourceType: configuration.Internal,
							Hosts:      "policy-engine.resource",
							Port:       9800,
						},
					},
				},

				fldPath: field.NewPath("resourceConnectInfo").Child("policy_engine"),
			},
			wantAllErrs: nil,
		},
		{
			name: "valid-external",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						PolicyEngine: &configuration.PolicyEngineInfo{
							SourceType: configuration.External,
							Hosts:      "policy-engine.resource",
							Port:       9800,
						},
					},
				},

				fldPath: field.NewPath("resourceConnectInfo").Child("policy_engine"),
			},
			wantAllErrs: nil,
		},
		{
			name: "invalid-external",
			args: args{
				c: &configuration.ClusterConfig{
					Proton_policy_engine: &configuration.ProtonDataConf{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						PolicyEngine: &configuration.PolicyEngineInfo{
							SourceType: configuration.External,
							Hosts:      "policy-engine.resource",
							Port:       9800,
						},
					},
				},

				fldPath: field.NewPath("resourceConnectInfo").Child("policy_engine"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-policy-engine",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("policy_engine"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-all",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						PolicyEngine: &configuration.PolicyEngineInfo{},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("policy_engine"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.policy_engine.hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.policy_engine.port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := validatePolicyEngine(tt.args.c.ResourceConnectInfo.PolicyEngine, tt.args.c, tt.args.fldPath)

			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("validatePolicyEngine() gotAllErrs != tt.wantAllErrs, %v", d)
			}
		})
	}

}

func Test_validateEtcd(t *testing.T) {
	type args struct {
		c       *configuration.ClusterConfig
		fldPath *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "valid",
			args: args{
				c: &configuration.ClusterConfig{
					Proton_etcd: &configuration.ProtonDataConf{},
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Etcd: &configuration.EtcdInfo{
							SourceType: configuration.Internal,
							Hosts:      "policy-engine.resource",
							Port:       9800,
							Secret:     "etcd-ssl",
						},
					},
				},

				fldPath: field.NewPath("resourceConnectInfo").Child("etcd"),
			},
			wantAllErrs: nil,
		},
		{
			name: "required-all",
			args: args{
				c: &configuration.ClusterConfig{
					ResourceConnectInfo: &configuration.ResourceConnectInfo{
						Etcd: &configuration.EtcdInfo{},
					},
				},
				fldPath: field.NewPath("resourceConnectInfo").Child("etcd"),
			},
			wantAllErrs: field.ErrorList{
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.etcd.hosts",
					BadValue: "",
				},
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "resourceConnectInfo.etcd.port",
					BadValue: 0,
					Detail:   "valid port: 0-65535",
				},
				{
					Type:     field.ErrorTypeRequired,
					Field:    "resourceConnectInfo.etcd.secret",
					BadValue: "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := validateEtcd(tt.args.c.ResourceConnectInfo.Etcd, tt.args.c, tt.args.fldPath)

			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("validateEtcd() gotAllErrs != tt.wantAllErrs, %v", d)
			}
		})
	}
}

func Test_isAssigned(t *testing.T) {

	tests := []struct {
		name string
		args *configuration.Auth
		want bool
	}{
		{
			name: "nil",
			want: false,
		},
		{
			name: "not-nil-but-init",
			args: &configuration.Auth{},
			want: false,
		},
		{
			name: "has-username",
			args: &configuration.Auth{
				Username: "test",
			},
			want: true,
		},
		{
			name: "has-password",
			args: &configuration.Auth{
				Password: "test",
			},
			want: true,
		},
		{
			name: "has-mechanism",
			args: &configuration.Auth{
				Mechanism: configuration.Plain,
			},
			want: true,
		},
		{
			name: "has-all",
			args: &configuration.Auth{
				Username:  "test",
				Password:  "test",
				Mechanism: configuration.Plain,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAssigned(tt.args)

			for _, d := range deep.Equal(got, tt.want) {
				t.Errorf("isAssigned() got != tt.want, %v", d)
			}
		})
	}
}
