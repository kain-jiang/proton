package request

import "component-manage/pkg/models/types"

type (
	PluginsKafka        = types.KafkaPluginConfig
	PluginsZookeeper    = types.ZookeeperPluginConfig
	PluginsOpensearch   = types.OpensearchPluginConfig
	PluginsRedis        = types.RedisPluginConfig
	PluginsETCD         = types.ETCDPluginConfig
	PluginsMariaDB      = types.MariaDBPluginConfig
	PluginsPolicyEngine = types.PolicyEnginePluginConfig
	PluginsMongoDB      = types.MongoDBPluginConfig
	PluginsNebula       = types.NebulaPluginConfig
	PluginsPrometheus   = types.PrometheusPluginConfig
)
