package opensearch

import (
	"fmt"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

const (
	SelfPluginName    = "opensearch"
	SelfPluginType    = "opensearch"
	SelfComponentType = "opensearch"
)

func EnableOpensearchPlugin(param types.OpensearchPluginConfig) error {
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
		return fmt.Errorf("set plugin opensearch error: %w", err)
	}
	return nil
}

func UpgradeOpensearchPlugin(param types.OpensearchPluginConfig) error {
	return EnableOpensearchPlugin(param)
}

func GetOpensearchPlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin opensearch not found", "")
	}
	return plugin, nil
}
