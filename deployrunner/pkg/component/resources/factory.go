package resources

import (
	"taskrunner/pkg/component"
	"taskrunner/trait"
	// load doc
)

// ReplaceApplicationComponentSchame replace the component schema
func ReplaceApplicationComponentSchame(cm *trait.ComponentMeta) *trait.Error {
	if cm.ComponentNode.ComponentDefineType != component.ComponentProtonResourceType {
		return nil
	}
	switch cm.Type {
	case RDSType:
		// cm.RawConfigSchema = _RdsSchema
	case REDISType:
		// cm.RawConfigSchema = _RedisSchema
	case MQType:
		// cm.RawConfigSchema = _MQSchema
	case OpensearchType:
		// cm.RawConfigSchema = _OpensearchConfigSchema
	case MongodbType:
		// cm.RawConfigSchema = _mongodbSChema
	case GraphType:
		cm.RawConfigSchema = _graphdbSChema
	default:
		return nil
	}
	return nil
}
