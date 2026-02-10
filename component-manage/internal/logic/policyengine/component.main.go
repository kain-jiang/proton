package policyengine

import (
	"fmt"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

func CreatePolicyEngine(name string, param *types.PolicyEngineComponentParams, eName string) (*types.ComponentPolicyEngine, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin policyengine not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt != nil {
		return nil, cerr.NewError(cerr.ComponentAlreadyExistsError, "component already exists", "")
	}

	// etcd不能被其他kafka使用
	opas, err := ListPolicyEngine()
	if err != nil {
		return nil, fmt.Errorf("list policy engine failed: %w", err)
	}

	usedETCDs := util.Map(opas, func(k *types.ComponentPolicyEngine) string {
		return k.Dependencies.ETCD
	})
	if util.InSlice(eName, usedETCDs) {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "please provide valid etcd", "etcd is already used")
	}

	// 获取etcd
	eCpt, err := global.Persist.GetComponentObject(eName)
	if err != nil {
		return nil, fmt.Errorf("get etcd error: %w", err)
	}
	if eCpt == nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "please provide valid etcd", "etcd not found")
	}
	eObj, err := eCpt.TryToETCD()
	if err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "provided etcd is invalid", err.Error())
	}

	// 检查参数（按需，将完善参数）
	if err := checkPolicyEngine(param, nil); err != nil {
		// return nil, fmt.Errorf("check policyengine params error: %w", err)
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check policyengine params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.PolicyEnginePluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := policyengineValues(param, eObj)
	err = preparePolicyEngine(param)
	if err != nil {
		return nil, fmt.Errorf("prepare policyengine error: %w", err)
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

	componentRelease := &types.ComponentPolicyEngine{
		Name:    name,
		Type:    SelfComponentType,
		Version: plugin.Version,
		Params:  param,
		Info:    generateInfo(name, param),
		Dependencies: &types.PolicyEngineComponentDependencies{
			ETCD: eName,
		},
	}

	// 持久化
	err = global.Persist.SetComponentObject(name, componentRelease.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return componentRelease, nil
}

func UpgradePolicyEngine(name string, param *types.PolicyEngineComponentParams) (*types.ComponentPolicyEngine, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin policyengine not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	policyengineObj, err := cpt.TryToPolicyEngine()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 获取etcd依赖
	eCpt, err := global.Persist.GetComponentObject(policyengineObj.Dependencies.ETCD)
	if err != nil {
		return nil, fmt.Errorf("get etcd component error: %w", err)
	}
	if eCpt == nil {
		return nil, fmt.Errorf("dependencies etcd component not found")
	}

	eObj, err := eCpt.TryToETCD()
	if err != nil {
		return nil, fmt.Errorf("parse etcd component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkPolicyEngine(param, policyengineObj.Params); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check policyengine params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.PolicyEnginePluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := policyengineValues(param, eObj)
	err = preparePolicyEngine(param)
	if err != nil {
		return nil, fmt.Errorf("prepare policyengine error: %w", err)
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

	policyengineObj.Params = param
	policyengineObj.Info = generateInfo(name, param)
	policyengineObj.Version = plugin.Version

	// 持久化
	err = global.Persist.SetComponentObject(name, policyengineObj.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return policyengineObj, nil
}

func DeletePolicyEngine(name string, toClean bool) (*types.ComponentPolicyEngine, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	policyengineObj, err := cpt.TryToPolicyEngine()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 如果需要清理存储，则需要wait等待
	_, err = global.HelmCli.NameSpace(policyengineObj.Params.Namespace).
		Uninstall(policyengineObj.Name, helm3.WithUninstallWait(toClean, 10*time.Minute), helm3.WithUninstallIgnoreNotFound(true))
	if err != nil {
		return nil, fmt.Errorf("uninstall component error: %w", err)
	}

	if toClean {
		err := clearPolicyEngine(policyengineObj.Params)
		if err != nil {
			return nil, fmt.Errorf("clear component error: %w", err)
		}
	}

	err = global.Persist.DelComponent(name)
	if err != nil {
		return nil, fmt.Errorf("delete component error: %w", err)
	}

	return policyengineObj, nil
}

func GetPolicyEngine(name string) (*types.ComponentPolicyEngine, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	policyengineObj, err := cpt.TryToPolicyEngine()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	return policyengineObj, nil
}

func ListPolicyEngine() ([]*types.ComponentPolicyEngine, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentPolicyEngine, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			policyengineObj, err := cpt.TryToPolicyEngine()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, policyengineObj)
		}
	}
	return result, nil
}
