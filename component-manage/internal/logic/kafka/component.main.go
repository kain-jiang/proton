package kafka

import (
	"fmt"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

func CreateKafka(name string, param *types.KafkaComponentParams, zkName string) (*types.ComponentKafka, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin kafka not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt != nil {
		return nil, cerr.NewError(cerr.ComponentAlreadyExistsError, "component already exists", "")
	}

	// zk不能被其他kafka使用
	kafkas, err := ListKafka()
	if err != nil {
		return nil, fmt.Errorf("list kafka failed: %w", err)
	}

	usedZKs := util.Map(kafkas, func(k *types.ComponentKafka) string {
		return k.Dependencies.Zookeeper
	})
	if util.InSlice(zkName, usedZKs) {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "please provide valid zookeeper", "zookeeper is already used")
	}

	// 获取zk
	zkCpt, err := global.Persist.GetComponentObject(zkName)
	if err != nil {
		return nil, fmt.Errorf("get zookeeper error: %w", err)
	}
	if zkCpt == nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "please provide valid zookeeper", "zookeeper not found")
	}
	zkObj, err := zkCpt.TryToZookeeper()
	if err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "provided zookeeper is invalid", err.Error())
	}

	// 检查参数（按需，将完善参数）
	if err := checkKafka(param, nil); err != nil {
		// return nil, fmt.Errorf("check kafka params error: %w", err)
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check kafka params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.KafkaPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := kafkaValues(param, zkObj)
	err = prepareKafka(param)
	if err != nil {
		return nil, fmt.Errorf("prepare kafka error: %w", err)
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

	componentRelease := &types.ComponentKafka{
		Name:    name,
		Type:    SelfComponentType,
		Version: plugin.Version,
		Params:  param,
		Info:    generateInfo(name, param),
		Dependencies: &types.KafkaComponentDependencies{
			Zookeeper: zkName,
		},
	}

	// 持久化
	err = global.Persist.SetComponentObject(name, componentRelease.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return componentRelease, nil
}

func UpgradeKafka(name string, param *types.KafkaComponentParams) (*types.ComponentKafka, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin kafka not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	kafkaObj, err := cpt.TryToKafka()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 获取zk依赖
	zkCpt, err := global.Persist.GetComponentObject(kafkaObj.Dependencies.Zookeeper)
	if err != nil {
		return nil, fmt.Errorf("get zookeeper component error: %w", err)
	}
	if zkCpt == nil {
		return nil, fmt.Errorf("dependencies zookeeper component not found")
	}

	zkObj, err := zkCpt.TryToZookeeper()
	if err != nil {
		return nil, fmt.Errorf("parse zookeeper component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkKafka(param, kafkaObj.Params); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check kafka params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.KafkaPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := kafkaValues(param, zkObj)
	err = prepareKafka(param)
	if err != nil {
		return nil, fmt.Errorf("prepare kafka error: %w", err)
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

	kafkaObj.Params = param
	kafkaObj.Info = generateInfo(name, param)
	kafkaObj.Version = plugin.Version

	// 持久化
	err = global.Persist.SetComponentObject(name, kafkaObj.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return kafkaObj, nil
}

func DeleteKafka(name string, toClean bool) (*types.ComponentKafka, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	kafkaObj, err := cpt.TryToKafka()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 如果需要清理存储，则需要wait等待
	_, err = global.HelmCli.NameSpace(kafkaObj.Params.Namespace).
		Uninstall(kafkaObj.Name, helm3.WithUninstallWait(toClean, 10*time.Minute), helm3.WithUninstallIgnoreNotFound(true))
	if err != nil {
		return nil, fmt.Errorf("uninstall component error: %w", err)
	}

	if toClean {
		err := clearKafka(kafkaObj.Params)
		if err != nil {
			return nil, fmt.Errorf("clear component error: %w", err)
		}
	}

	err = global.Persist.DelComponent(name)
	if err != nil {
		return nil, fmt.Errorf("delete component error: %w", err)
	}

	return kafkaObj, nil
}

func GetKafka(name string) (*types.ComponentKafka, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	kafkaObj, err := cpt.TryToKafka()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	return kafkaObj, nil
}

func ListKafka() ([]*types.ComponentKafka, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentKafka, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			kafkaObj, err := cpt.TryToKafka()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, kafkaObj)
		}
	}
	return result, nil
}
