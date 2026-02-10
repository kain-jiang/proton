package response

import (
	"component-manage/pkg/models/types"
)

type (
	Component             = types.ComponentObject
	ComponentKafka        = types.ComponentKafka
	ComponentZookeeper    = types.ComponentZookeeper
	ComponentOpensearch   = types.ComponentOpensearch
	ComponentRedis        = types.ComponentRedis
	ComponentETCD         = types.ComponentETCD
	ComponentMariaDB      = types.ComponentMariaDB
	ComponentPolicyEngine = types.ComponentPolicyEngine
	ComponentMongoDB      = types.ComponentMongoDB
	ComponentNebula       = types.ComponentNebula
	ComponentPrometheus   = types.ComponentPrometheus
)
