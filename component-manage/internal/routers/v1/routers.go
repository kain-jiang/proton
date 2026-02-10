package v1

import (
	"component-manage/internal/routers/v1/component"
	"component-manage/internal/routers/v1/health"
	"component-manage/internal/routers/v1/plugin"

	"github.com/gin-gonic/gin"
)

func RegistryApi(e *gin.Engine) error {
	e.GET("/health/alive", health.ApiHealthAlive)
	e.GET("/health/ready", health.ApiHealthReady)

	apiv1 := e.Group("/api/component-manage/v1")
	{
		{
			apiv1.POST("/components/plugin/kafka", plugin.ApiPluginKafkaEnable)
			apiv1.PUT("/components/plugin/kafka", plugin.ApiPluginKafkaUpgrade)
			apiv1.GET("/components/plugin/kafka", plugin.ApiPluginKafkaGet)
		}
		{
			apiv1.POST("/components/plugin/zookeeper", plugin.ApiPluginZookeeperEnable)
			apiv1.PUT("/components/plugin/zookeeper", plugin.ApiPluginZookeeperUpgrade)
			apiv1.GET("/components/plugin/zookeeper", plugin.ApiPluginZookeeperGet)
		}
		{
			apiv1.POST("/components/plugin/opensearch", plugin.ApiPluginOpensearchEnable)
			apiv1.PUT("/components/plugin/opensearch", plugin.ApiPluginOpensearchUpgrade)
			apiv1.GET("/components/plugin/opensearch", plugin.ApiPluginOpensearchGet)
		}
		{
			apiv1.POST("/components/plugin/redis", plugin.ApiPluginRedisEnable)
			apiv1.PUT("/components/plugin/redis", plugin.ApiPluginRedisUpgrade)
			apiv1.GET("/components/plugin/redis", plugin.ApiPluginRedisGet)
		}
		{
			apiv1.POST("/components/plugin/etcd", plugin.ApiPluginETCDEnable)
			apiv1.PUT("/components/plugin/etcd", plugin.ApiPluginETCDUpgrade)
			apiv1.GET("/components/plugin/etcd", plugin.ApiPluginETCDGet)
		}
		{
			apiv1.POST("/components/plugin/mariadb", plugin.ApiPluginMariaDBEnable)
			apiv1.PUT("/components/plugin/mariadb", plugin.ApiPluginMariaDBUpgrade)
			apiv1.GET("/components/plugin/mariadb", plugin.ApiPluginMariaDBGet)
		}
		{
			apiv1.POST("/components/plugin/policyengine", plugin.ApiPluginPolicyEngineEnable)
			apiv1.PUT("/components/plugin/policyengine", plugin.ApiPluginPolicyEngineUpgrade)
			apiv1.GET("/components/plugin/policyengine", plugin.ApiPluginPolicyEngineGet)
		}
		{
			apiv1.POST("/components/plugin/mongodb", plugin.ApiPluginMongoDBEnable)
			apiv1.PUT("/components/plugin/mongodb", plugin.ApiPluginMongoDBUpgrade)
			apiv1.GET("/components/plugin/mongodb", plugin.ApiPluginMongoDBGet)
		}
		{
			apiv1.POST("/components/plugin/nebula", plugin.ApiPluginNebulaEnable)
			apiv1.PUT("/components/plugin/nebula", plugin.ApiPluginNebulaUpgrade)
			apiv1.GET("/components/plugin/nebula", plugin.ApiPluginNebulaGet)
		}
		{
			apiv1.POST("/components/plugin/prometheus", plugin.ApiPluginPrometheusEnable)
			apiv1.PUT("/components/plugin/prometheus", plugin.ApiPluginPrometheusUpgrade)
			apiv1.GET("/components/plugin/prometheus", plugin.ApiPluginPrometheusGet)
		}
	}

	{
		apiv1.GET("/components/release/all", component.ApiComponentAllList)

		{
			apiv1.POST("/components/release/kafka/:name", component.ApiComponentKafkaCreate)
			apiv1.PUT("/components/release/kafka/:name", component.ApiComponentKafkaUpgrade)
			apiv1.GET("/components/release/kafka/:name", component.ApiComponentKafkaGet)
			apiv1.GET("/components/release/kafka", component.ApiComponentKafkaList)
			apiv1.DELETE("/components/release/kafka/:name", component.ApiComponentKafkaDelete)
		}
		{
			apiv1.POST("/components/release/zookeeper/:name", component.ApiComponentZookeeperCreate)
			apiv1.PUT("/components/release/zookeeper/:name", component.ApiComponentZookeeperUpgrade)
			apiv1.GET("/components/release/zookeeper/:name", component.ApiComponentZookeeperGet)
			apiv1.GET("/components/release/zookeeper", component.ApiComponentZookeeperList)
			apiv1.DELETE("/components/release/zookeeper/:name", component.ApiComponentZookeeperDelete)
		}
		{
			apiv1.POST("/components/release/opensearch/:name", component.ApiComponentOpensearchCreate)
			apiv1.PUT("/components/release/opensearch/:name", component.ApiComponentOpensearchUpgrade)
			apiv1.GET("/components/release/opensearch/:name", component.ApiComponentOpensearchGet)
			apiv1.GET("/components/release/opensearch", component.ApiComponentOpensearchList)
			apiv1.DELETE("/components/release/opensearch/:name", component.ApiComponentOpensearchDelete)
		}
		{
			apiv1.POST("/components/release/redis/:name", component.ApiComponentRedisCreate)
			apiv1.PUT("/components/release/redis/:name", component.ApiComponentRedisUpgrade)
			apiv1.GET("/components/release/redis/:name", component.ApiComponentRedisGet)
			apiv1.GET("/components/release/redis", component.ApiComponentRedisList)
			apiv1.DELETE("/components/release/redis/:name", component.ApiComponentRedisDelete)
		}
		{
			apiv1.POST("/components/release/etcd/:name", component.ApiComponentETCDCreate)
			apiv1.PUT("/components/release/etcd/:name", component.ApiComponentETCDUpgrade)
			apiv1.GET("/components/release/etcd/:name", component.ApiComponentETCDGet)
			apiv1.GET("/components/release/etcd", component.ApiComponentETCDList)
			apiv1.DELETE("/components/release/etcd/:name", component.ApiComponentETCDDelete)
		}
		{
			apiv1.POST("/components/release/mariadb/:name", component.ApiComponentMariaDBCreate)
			apiv1.PUT("/components/release/mariadb/:name", component.ApiComponentMariaDBUpgrade)
			apiv1.GET("/components/release/mariadb/:name", component.ApiComponentMariaDBGet)
			apiv1.GET("/components/release/mariadb", component.ApiComponentMariaDBList)
			apiv1.DELETE("/components/release/mariadb/:name", component.ApiComponentMariaDBDelete)
		}
		{
			apiv1.POST("/components/release/policyengine/:name", component.ApiComponentPolicyEngineCreate)
			apiv1.PUT("/components/release/policyengine/:name", component.ApiComponentPolicyEngineUpgrade)
			apiv1.GET("/components/release/policyengine/:name", component.ApiComponentPolicyEngineGet)
			apiv1.GET("/components/release/policyengine/", component.ApiComponentPolicyEngineList)
			apiv1.DELETE("/components/release/policyengine/:name", component.ApiComponentPolicyEngineDelete)
		}
		{
			apiv1.POST("/components/release/mongodb/:name", component.ApiComponentMongoDBCreate)
			apiv1.PUT("/components/release/mongodb/:name", component.ApiComponentMongoDBUpgrade)
			apiv1.GET("/components/release/mongodb/:name", component.ApiComponentMongoDBGet)
			apiv1.GET("/components/release/mongodb", component.ApiComponentMongoDBList)
			apiv1.DELETE("/components/release/mongodb/:name", component.ApiComponentMongoDBDelete)
		}
		{
			apiv1.POST("/components/release/nebula/:name", component.ApiComponentNebulaCreate)
			apiv1.PUT("/components/release/nebula/:name", component.ApiComponentNebulaUpgrade)
			apiv1.GET("/components/release/nebula/:name", component.ApiComponentNebulaGet)
			apiv1.GET("/components/release/nebula", component.ApiComponentNebulaList)
			apiv1.DELETE("/components/release/nebula/:name", component.ApiComponentNebulaDelete)
		}
		{
			apiv1.POST("/components/release/prometheus/:name", component.ApiComponentPrometheusCreate)
			apiv1.PUT("/components/release/prometheus/:name", component.ApiComponentPrometheusUpgrade)
			apiv1.GET("/components/release/prometheus/:name", component.ApiComponentPrometheusGet)
			apiv1.GET("/components/release/prometheus", component.ApiComponentPrometheusList)
			apiv1.DELETE("/components/release/prometheus/:name", component.ApiComponentPrometheusDelete)
		}

	}

	return nil
}
