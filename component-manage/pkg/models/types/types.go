package types

import (
	"component-manage/pkg/models/types/components"
	"component-manage/pkg/models/types/plugins"
)

type (
	Resource             = components.Resources
	ResourceRequirements = components.ResourceRequirements

	ComponentObject       = components.ComponentObject
	ComponentKafka        = components.ComponentKafka
	ComponentZookeeper    = components.ComponentZookeeper
	ComponentOpensearch   = components.ComponentOpensearch
	ComponentRedis        = components.ComponentRedis
	ComponentETCD         = components.ComponentETCD
	ComponentMariaDB      = components.ComponentMariaDB
	ComponentPolicyEngine = components.ComponentPolicyEngine
	ComponentMongoDB      = components.ComponentMongoDB
	ComponentNebula       = components.ComponentNebula
	ComponentPrometheus   = components.ComponentPrometheus

	KafkaComponentInfo         = components.KafkaComponentInfo
	KafkaComponentParams       = components.KafkaComponentParams
	KafkaComponentDependencies = components.KafkaComponentDependencies
	KafkaAuth                  = components.KafkaAuth

	ZookeeperComponentParams = components.ZookeeperComponentParams
	ZookeeperComponentInfo   = components.ZookeeperComponentInfo
	ZookeeperSASL            = components.ZookeeperSASL

	OpensearchComponentParams = components.OpensearchComponentParams
	OpensearchComponentInfo   = components.OpensearchComponentInfo

	RedisComponentParams = components.RedisComponentParams
	RedisComponentInfo   = components.RedisComponentInfo

	ETCDComponentParams = components.ETCDComponentParams
	ETCDComponentInfo   = components.ETCDComponentInfo
	ETCDCAInfo          = components.ETCDCAInfo

	MariaDBComponentParams = components.MariaDBComponentParams
	MariaDBComponentInfo   = components.MariaDBComponentInfo

	MongoDBComponentParams = components.MongoDBComponentParams
	MongoDBComponentInfo   = components.MongoDBComponentInfo

	PolicyEngineComponentParams       = components.PolicyEngineComponentParams
	PolicyEngineComponentInfo         = components.PolicyEngineComponentInfo
	PolicyEngineComponentDependencies = components.PolicyEngineComponentDependencies

	NebulaComponentParams = components.NebulaComponentParams
	NebulaComponentInfo   = components.NebulaComponentInfo

	PrometheusComponentParams = components.PrometheusComponentParams
	PrometheusComponentInfo   = components.PrometheusComponentInfo

	PluginObject             = plugins.PluginObject
	KafkaPluginConfig        = plugins.KafkaPluginConfig
	ZookeeperPluginConfig    = plugins.ZookeeperPluginConfig
	OpensearchPluginConfig   = plugins.OpensearchPluginConfig
	RedisPluginConfig        = plugins.RedisPluginConfig
	ETCDPluginConfig         = plugins.ETCDPluginConfig
	MariaDBPluginConfig      = plugins.MariaDBPluginConfig
	PolicyEnginePluginConfig = plugins.PolicyEnginePluginConfig
	MongoDBPluginConfig      = plugins.MongoDBPluginConfig
	NebulaPluginConfig       = plugins.NebulaPluginConfig
	PrometheusPluginConfig   = plugins.PrometheusPluginConfig
)
