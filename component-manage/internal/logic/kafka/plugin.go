package kafka

import (
	"fmt"

	"component-manage/pkg/util"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/types"
)

const (
	SelfPluginName    = "kafka"
	SelfPluginType    = "kafka"
	SelfComponentType = "kafka"
)

// EnableKafkaPlugin 启用 kafka 插件
func EnableKafkaPlugin(param types.KafkaPluginConfig) error {
	// params 转 map
	m, err := util.ToMap(param)
	if err != nil {
		return fmt.Errorf("params to map error: %w", err)
	}

	pluginRelease := &types.PluginObject{
		Name:    SelfPluginName,
		Type:    SelfPluginType,
		Version: param.ChartVersion,
		Config:  m,
	}

	if err := global.Persist.SetPluginObject(SelfPluginName, pluginRelease); err != nil {
		return fmt.Errorf("set plugin kafka error: %w", err)
	}
	return nil
}

// UpgradeKafkaPlugin 升级 kafka 插件
func UpgradeKafkaPlugin(param types.KafkaPluginConfig) error {
	return EnableKafkaPlugin(param)
}

// GetKafkaPlugin 获取 kafka 插件
func GetKafkaPlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin kafka not found", "")
	}
	return plugin, nil
}
