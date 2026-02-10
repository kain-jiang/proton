package redis

import (
	"fmt"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

func CreateRedis(name string, param *types.RedisComponentParams) (*types.ComponentRedis, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin redis not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt != nil {
		return nil, cerr.NewError(cerr.ComponentAlreadyExistsError, "component already exists", "")
	}

	// 检查参数（按需，将完善参数）
	if err := checkRedis(param, nil); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check redis params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.RedisPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := RedisValues(param)
	err = prepareRedis(param)
	if err != nil {
		return nil, fmt.Errorf("prepare redis error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		name, cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValues(values),
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade or install error: %w", err)
	}

	componentRelease := &types.ComponentRedis{
		Name:    name,
		Type:    SelfComponentType,
		Version: plugin.Version,
		Params:  param,
		Info:    generateInfo(name, param),
	}

	// 持久化
	err = global.Persist.SetComponentObject(name, componentRelease.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return componentRelease, nil
}

func UpgradeRedis(name string, param *types.RedisComponentParams) (*types.ComponentRedis, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin redis not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	rObj, err := cpt.TryToRedis()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkRedis(param, rObj.Params); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check redis params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.RedisPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := RedisValues(param)
	err = prepareRedis(param)
	if err != nil {
		return nil, fmt.Errorf("prepare redis error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		name, cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValues(values),
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade or install error: %w", err)
	}
	rObj.Params = param
	rObj.Info = generateInfo(name, param)
	rObj.Version = plugin.Version

	// 持久化
	err = global.Persist.SetComponentObject(name, rObj.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return rObj, nil
}

func DeleteRedis(name string, toClean bool) (*types.ComponentRedis, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	rObj, err := cpt.TryToRedis()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 如果需要清理存储，则需要wait等待
	_, err = global.HelmCli.NameSpace(rObj.Params.Namespace).
		Uninstall(rObj.Name, helm3.WithUninstallWait(toClean, 10*time.Minute), helm3.WithUninstallIgnoreNotFound(true))
	if err != nil {
		return nil, fmt.Errorf("uninstall component error: %w", err)
	}

	if toClean {
		err := clearRedis(rObj.Params)
		if err != nil {
			return nil, fmt.Errorf("clear component error: %w", err)
		}
	}

	err = global.Persist.DelComponent(name)
	if err != nil {
		return nil, fmt.Errorf("delete component error: %w", err)
	}

	return rObj, nil
}

func GetRedis(name string) (*types.ComponentRedis, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	rObj, err := cpt.TryToRedis()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	return rObj, nil
}

func ListRedis() ([]*types.ComponentRedis, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentRedis, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			rObj, err := cpt.TryToRedis()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, rObj)
		}
	}
	return result, nil
}
