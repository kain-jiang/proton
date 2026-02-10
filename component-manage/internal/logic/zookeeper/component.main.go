package zookeeper

import (
	"fmt"
	"reflect"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

func CreateZookeeper(name string, param *types.ZookeeperComponentParams) (*types.ComponentZookeeper, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin zookeeper not found", "")
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
	if err := checkZookeeper(param, nil); err != nil {
		// return nil, fmt.Errorf("check zookeeper params error: %w", err)
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check zookeeper params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.ZookeeperPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := zookeeperValues(param)
	err = prepareZookeeper(param)
	if err != nil {
		return nil, fmt.Errorf("prepare zookeeper error: %w", err)
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

	componentRelease := &types.ComponentZookeeper{
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

func UpgradeZookeeper(name string, param *types.ZookeeperComponentParams) (*types.ComponentZookeeper, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin zookeeper not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	zkObj, err := cpt.TryToZookeeper()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkZookeeper(param, zkObj.Params); err != nil {
		// return nil, fmt.Errorf("check zookeeper params error: %w", err)
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check zookeeper params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.ZookeeperPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := zookeeperValues(param)
	err = prepareZookeeper(param)
	if err != nil {
		return nil, fmt.Errorf("prepare zookeeper error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		name, cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeRecreatePods(true),
		helm3.WithUpgradeValues(values),
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade or install error: %w", err)
	}

	zkObj.Params = param
	newInfo := generateInfo(name, param)

	// 如何判断两个结构体字段是否一致？
	if !reflect.DeepEqual(zkObj.Info, newInfo) {
		// 发生了变更，依赖需要更新
		global.Logger.WithField("component", name).WithField("info", newInfo).Warn("component info is changed")
	}

	zkObj.Info = newInfo
	zkObj.Version = plugin.Version
	// 持久化
	err = global.Persist.SetComponentObject(name, zkObj.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return zkObj, nil
}

func DeleteZookeeper(name string, toClean bool) (*types.ComponentZookeeper, error) {
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component object error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}

	cptObj, err := cpt.TryToZookeeper()
	if err != nil {
		return nil, fmt.Errorf("try to zookeeper error: %w", err)
	}

	// 如果要清理数据则需要等待完成
	_, err = global.HelmCli.NameSpace(cptObj.Params.Namespace).
		Uninstall(cptObj.Name, helm3.WithUninstallWait(toClean, 10*time.Minute), helm3.WithUninstallIgnoreNotFound(true))
	if err != nil {
		return nil, fmt.Errorf("helm uninstall error: %w", err)
	}

	if toClean {
		err := clearZookeeper(cptObj.Params)
		if err != nil {
			return nil, fmt.Errorf("clear component error: %w", err)
		}
	}

	err = global.Persist.DelComponent(name)
	if err != nil {
		return nil, fmt.Errorf("delete component error: %w", err)
	}

	return cptObj, nil
}

func GetZookeeper(name string) (*types.ComponentZookeeper, error) {
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	zkObj, err := cpt.TryToZookeeper()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	return zkObj, nil
}

func ListZookeeper() ([]*types.ComponentZookeeper, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentZookeeper, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			zkObj, err := cpt.TryToZookeeper()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, zkObj)
		}
	}
	return result, nil
}
