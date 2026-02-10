package mongodb

import (
	"fmt"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

const (
	SelfPluginName    = "mongodb"
	SelfPluginType    = "mongodb"
	SelfComponentType = "mongodb"

	operatorName = "mongodb-operator"
)

func EnableMongoDBPlugin(param types.MongoDBPluginConfig) error {
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
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("install mongodb operator failed: %w", err)
	}

	m, err := util.ToMap(param)
	if err != nil {
		return fmt.Errorf("params to map error: %w", err)
	}

	pluginRelease := &types.PluginObject{
		Name:    SelfPluginName,
		Type:    SelfPluginType,
		Version: fmt.Sprintf("%s+%s", param.ChartVersion, base.ImageTag(param.Images.MongoDB)),
		Config:  m,
	}

	if err := global.Persist.SetPluginObject(SelfPluginName, pluginRelease); err != nil {
		return fmt.Errorf("set plugin mongodb error: %w", err)
	}
	return nil
}

func UpgradeMongoDBPlugin(param types.MongoDBPluginConfig) error {
	return EnableMongoDBPlugin(param)
}

func GetMongoDBPlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin mongodb not found", "")
	}
	return plugin, nil
}
