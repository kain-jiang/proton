package zookeeper

import (
	"fmt"

	"component-manage/pkg/util"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/types"
)

const (
	SelfPluginName    = "zookeeper"
	SelfPluginType    = "zookeeper"
	SelfComponentType = "zookeeper"
)

// EnableZookeeperPlugin 启用 zookeeper 插件
func EnableZookeeperPlugin(param types.ZookeeperPluginConfig) error {
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
		return fmt.Errorf("set plugin zookeeper error: %w", err)
	}
	return nil
}

// UpgradeZookeeperPlugin 升级 zookeeper 插件
func UpgradeZookeeperPlugin(param types.ZookeeperPluginConfig) error {
	return EnableZookeeperPlugin(param)
}

// GetZookeeperPlugin 获取 zookeeper 插件
func GetZookeeperPlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin zookeeper not found", "")
	}
	return plugin, nil
}
