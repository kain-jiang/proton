package prometheus

import (
	"fmt"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

func CreatePrometheus(name string, param *types.PrometheusComponentParams, etcdName string) (*types.ComponentPrometheus, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin prometheus not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt != nil {
		return nil, cerr.NewError(cerr.ComponentAlreadyExistsError, "component already exists", "")
	}

	// 获取etcd
	etcdCpt, err := global.Persist.GetComponentObject(etcdName)
	if err != nil {
		return nil, fmt.Errorf("get proton etcd error: %w", err)
	}
	if etcdCpt == nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "please provide valid proton etcd", "proton etcd not found")
	}
	etcdObj, err := etcdCpt.TryToETCD()
	if err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "provided proton etcd is invalid", err.Error())
	}

	// 检查参数（按需，将完善参数）
	if err := checkPrometheus(param, nil, etcdObj); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check prometheus params error", err.Error())
	}

	// 准备 prometheus 使用的 etcd 证书
	certs, err := prepareEtcdCertForPrometheus(name, param)
	if err != nil {
		return nil, fmt.Errorf("prepare etcd certs failed: %w", err)
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.PrometheusPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := PrometheusValues(param, certs)

	err = preparePrometheus(param)
	if err != nil {
		return nil, fmt.Errorf("prepare prometheus error: %w", err)
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

	componentRelease := &types.ComponentPrometheus{
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

func UpgradePrometheus(name string, param *types.PrometheusComponentParams) (*types.ComponentPrometheus, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin prometheus not found", "")
	}

	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	rObj, err := cpt.TryToPrometheus()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkPrometheus(param, rObj.Params, nil); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check prometheus params error", err.Error())
	}

	// 准备 prometheus 使用的 etcd 证书
	certs, err := prepareEtcdCertForPrometheus(name, param)
	if err != nil {
		return nil, fmt.Errorf("prepare etcd certs failed: %w", err)
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.PrometheusPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := PrometheusValues(param, certs)
	err = preparePrometheus(param)
	if err != nil {
		return nil, fmt.Errorf("prepare prometheus error: %w", err)
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

func DeletePrometheus(name string, toClean bool) (*types.ComponentPrometheus, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	rObj, err := cpt.TryToPrometheus()
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
		err := clearPrometheus(rObj.Params)
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

func GetPrometheus(name string) (*types.ComponentPrometheus, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}

	rObj, err := cpt.TryToPrometheus()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	return rObj, nil
}

func ListPrometheus() ([]*types.ComponentPrometheus, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentPrometheus, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			rObj, err := cpt.TryToPrometheus()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, rObj)
		}
	}
	return result, nil
}
