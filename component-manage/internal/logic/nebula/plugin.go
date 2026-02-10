package nebula

import (
	"fmt"
	"time"

	"component-manage/pkg/helm3"
	"component-manage/pkg/util"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/types"
)

const (
	SelfPluginName    = "nebula"
	SelfPluginType    = "nebula"
	SelfComponentType = "nebula"

	operatorName = "nebula-operator"
)

// EnableNebulaPlugin 启用 nebula 插件
func EnableNebulaPlugin(param types.NebulaPluginConfig) error {
	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), param.ChartName, param.ChartVersion)
	if err != nil {
		return fmt.Errorf("fetch chart error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		operatorName, cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeWait(true, 10*time.Minute),
		helm3.WithUpgradeValues(helm3.M{
			"controllerManager": helm3.M{
				"image": helm3.M{
					"registry": global.Config.Config.Registry,
				},
				"replicas": 1,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("install nebula operator failed: %w", err)
	}

	m, err := util.ToMap(param)
	if err != nil {
		return fmt.Errorf("params to map error: %w", err)
	}

	pluginRelease := &types.PluginObject{
		Name:    SelfPluginName,
		Type:    SelfPluginType,
		Version: fmt.Sprintf("%s+%s", param.ChartVersion, base.ImageTag(param.Images.GraphD)),
		Config:  m,
	}

	if err := global.Persist.SetPluginObject(SelfPluginName, pluginRelease); err != nil {
		return fmt.Errorf("set plugin nebula error: %w", err)
	}
	return nil
}

// UpgradeNebulaPlugin 升级 nebula 插件
func UpgradeNebulaPlugin(param types.NebulaPluginConfig) error {
	return EnableNebulaPlugin(param)
}

func GetNebulaPlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin nebula not found", "")
	}
	return plugin, nil
}
