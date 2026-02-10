package opensearch

import (
	"fmt"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

func CreateOpensearch(name string, param *types.OpensearchComponentParams) (*types.ComponentOpensearch, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin opensearch not found", "")
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
	if err := checkOpensearch(param, nil); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check opensearch params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.OpensearchPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := OpensearchValues(name, param)
	err = prepareOpensearch(param)
	if err != nil {
		return nil, fmt.Errorf("prepare opensearch error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		nameForMaster(name), cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValuesAny(values), // TODO 不标准 map[string]any
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade or install error: %w", err)
	}

	componentRelease := &types.ComponentOpensearch{
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

func UpgradeOpensearch(name string, param *types.OpensearchComponentParams) (*types.ComponentOpensearch, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin opensearch not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	oObj, err := cpt.TryToOpensearch()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkOpensearch(param, oObj.Params); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check opensearch params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.OpensearchPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := OpensearchValues(name, param)
	err = prepareOpensearch(param)
	if err != nil {
		return nil, fmt.Errorf("prepare opensearch error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		nameForMaster(name), cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValuesAny(values), // TODO 不标准 map[string]any
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade or install error: %w", err)
	}
	oObj.Params = param
	oObj.Info = generateInfo(name, param)
	oObj.Version = plugin.Version

	// 持久化
	err = global.Persist.SetComponentObject(name, oObj.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return oObj, nil
}

func DeleteOpensearch(name string, toClean bool) (*types.ComponentOpensearch, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	oObj, err := cpt.TryToOpensearch()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 如果需要清理存储，则需要wait等待
	_, err = global.HelmCli.NameSpace(oObj.Params.Namespace).
		Uninstall(
			nameForMaster(name),
			helm3.WithUninstallWait(toClean, 10*time.Minute),
			helm3.WithUninstallIgnoreNotFound(true),
		)
	if err != nil {
		return nil, fmt.Errorf("uninstall component error: %w", err)
	}

	if toClean {
		err := clearOpensearch(oObj.Params)
		if err != nil {
			return nil, fmt.Errorf("clear component error: %w", err)
		}
	}

	err = global.Persist.DelComponent(name)
	if err != nil {
		return nil, fmt.Errorf("delete component error: %w", err)
	}

	return oObj, nil
}

func GetOpensearch(name string) (*types.ComponentOpensearch, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	oObj, err := cpt.TryToOpensearch()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	return oObj, nil
}

func ListOpensearch() ([]*types.ComponentOpensearch, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentOpensearch, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			oObj, err := cpt.TryToOpensearch()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, oObj)
		}
	}
	return result, nil
}
