package mariadb

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
	SelfPluginName    = "mariadb"
	SelfPluginType    = "mariadb"
	SelfComponentType = "mariadb"

	operatorName = "rds-mariadb-operator"
)

func EnableMariaDBPlugin(param types.MariaDBPluginConfig) error {
	err := checkMariaDBPlugin(param)
	if err != nil {
		return cerr.NewError(cerr.ParamsInvalidError, "check mariadb param failed", err.Error())
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), param.ChartName, param.ChartVersion)
	if err != nil {
		return fmt.Errorf("fetch chart error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		operatorName, cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeWait(true, 10*time.Minute),
		helm3.WithUpgradeValues(helm3.M{
			"image": helm3.M{
				"controller": helm3.M{},
				"proxy":      helm3.M{},
				"registry":   global.Config.Config.Registry,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("install mariadb operator failed: %w", err)
	}

	m, err := util.ToMap(param)
	if err != nil {
		return fmt.Errorf("params to map error: %w", err)
	}

	pluginRelease := &types.PluginObject{
		Name:    SelfPluginName,
		Type:    SelfPluginType,
		Version: fmt.Sprintf("%s+%s", param.ChartVersion, base.ImageTag(param.Images.MariaDB)),
		Config:  m,
	}

	if err := global.Persist.SetPluginObject(SelfPluginName, pluginRelease); err != nil {
		return fmt.Errorf("set plugin mariadb error: %w", err)
	}
	return nil
}

func UpgradeMariaDBPlugin(param types.MariaDBPluginConfig) error {
	return EnableMariaDBPlugin(param)
}

func GetMariaDBPlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin mariadb not found", "")
	}
	return plugin, nil
}

func checkMariaDBPlugin(param types.MariaDBPluginConfig) error {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return fmt.Errorf("get plugin error: %w", err)
	}
	if plugin != nil {
		// 升级场景
		if plgInfo, err := util.FromMap[types.MariaDBPluginConfig](plugin.Config); err != nil {
			return fmt.Errorf("parse plugin error: %w", err)
		} else {
			nowTag := base.ImageTag(param.Images.MariaDB)
			oldTag := base.ImageTag(plgInfo.Images.MariaDB)
			if util.VersionOrdinal(nowTag) < util.VersionOrdinal(oldTag) {
				return fmt.Errorf("mariadb image cannot be downgraded: from %s to %s", oldTag, nowTag)
			}
		}
	}
	return nil
}
