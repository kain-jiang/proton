package etcd

import (
	"fmt"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

const (
	SelfPluginName    = "etcd"
	SelfPluginType    = "etcd"
	SelfComponentType = "etcd"
)

func EnableETCDPlugin(param types.ETCDPluginConfig) error {
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
		return fmt.Errorf("set plugin etcd error: %w", err)
	}
	return nil
}

func UpgradeETCDPlugin(param types.ETCDPluginConfig) error {
	return EnableETCDPlugin(param)
}

func GetETCDPlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin etcd not found", "")
	}
	return plugin, nil
}
