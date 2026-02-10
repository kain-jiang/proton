package request

import (
	"component-manage/pkg/models/types"
)

type ComponentKafka struct {
	Params       *types.KafkaComponentParams `json:"params" binding:"required"`
	Dependencies struct {
		Zookeeper string `json:"zookeeper" binding:"required"`
	} `json:"dependencies" binding:"required"`
}

type ComponentZookeeper struct {
	Params *types.ZookeeperComponentParams `json:"params" binding:"required"`
}

type ComponentOpensearch struct {
	Params *types.OpensearchComponentParams `json:"params" binding:"required"`
}

type ComponentRedis struct {
	Params *types.RedisComponentParams `json:"params" binding:"required"`
}

type ComponentETCD struct {
	Params *types.ETCDComponentParams `json:"params" binding:"required"`
}

type ComponentMariaDB struct {
	Params *types.MariaDBComponentParams `json:"params" binding:"required"`
}

type ComponentPolicyEngine struct {
	Params       *types.PolicyEngineComponentParams `json:"params" binding:"required"`
	Dependencies struct {
		ETCD string `json:"etcd" binding:"required"`
	} `json:"dependencies" binding:"required"`
}
type ComponentMongoDB struct {
	Params *types.MongoDBComponentParams `json:"params" binding:"required"`
}

type ComponentNebula struct {
	Params *types.NebulaComponentParams `json:"params" binding:"required"`
}

type ComponentPrometheus struct {
	Params       *types.PrometheusComponentParams `json:"params" binding:"required"`
	Dependencies struct {
		ETCD string `json:"etcd" binding:"required"`
	} `json:"dependencies" binding:"required"`
}
