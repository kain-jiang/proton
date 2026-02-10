package etcd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"component-manage/internal/global"
	"component-manage/internal/logic/base"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/helm3"
	"component-manage/pkg/models/types"
	"component-manage/pkg/models/types/components"
	"component-manage/pkg/util"
)

// GetETCDEndpointsInKubernetes 返回在 kubernetes 中访问 etcd 所用的 endpoint 列表
func GetETCDEndpointsInKubernetes(e *components.ComponentETCD) (endpoints []string) {
	replicas := len(e.Params.Hosts)
	if replicas == 0 {
		replicas = e.Params.ReplicaCount
	}
	for i := 0; i < replicas; i++ {
		podDNSName := ""
		podDNSName = strings.Join([]string{
			fmt.Sprintf("%s-%s", base.TemplateName(e.Name, "proton-etcd"), strconv.Itoa(i)),
			fmt.Sprintf("%s-headless", base.TemplateName(e.Name, "proton-etcd")),
			e.Params.Namespace, "svc", "cluster.local",
		}, ".")
		u := url.URL{
			Scheme: "https",
			Host:   net.JoinHostPort(podDNSName, strconv.Itoa(2379)),
		}
		endpoints = append(endpoints, u.String())
	}
	return
}

func CreateETCD(name string, param *types.ETCDComponentParams) (*types.ComponentETCD, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin etcd not found", "")
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
	if err := checkETCD(param, nil, name); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check etcd params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.ETCDPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(global.HelmCli, global.Config.ConfigChartmuseumToRepoEntry(), global.Config.ConfigOCIRegistryInfo(), plgInfo.ChartName, plgInfo.ChartVersion)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := ETCDValues(param, name)
	err = prepareETCD(param)
	if err != nil {
		return nil, fmt.Errorf("prepare etcd error: %w", err)
	}

	// 安装前创建证书secret
	if err := generateCert(name, param); err != nil {
		return nil, fmt.Errorf("generate etcd cert error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		name, cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValues(values),
		helm3.WithUpgradeRecreatePods(true),
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade or install error: %w", err)
	}

	componentRelease := &types.ComponentETCD{
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

func UpgradeETCD(name string, param *types.ETCDComponentParams) (*types.ComponentETCD, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin etcd not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	eObj, err := cpt.TryToETCD()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkETCD(param, eObj.Params, name); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check etcd params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.ETCDPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	cht, err := helm3.FetchChart(
		global.HelmCli,
		global.Config.ConfigChartmuseumToRepoEntry(),
		global.Config.ConfigOCIRegistryInfo(),
		plgInfo.ChartName, plgInfo.ChartVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch chart error: %w", err)
	}

	values := ETCDValues(param, name)
	err = prepareETCD(param)
	if err != nil {
		return nil, fmt.Errorf("prepare etcd error: %w", err)
	}

	// 安装前创建证书secret
	if err := generateCert(name, param); err != nil {
		return nil, fmt.Errorf("generate etcd cert error: %w", err)
	}

	_, err = global.HelmCli.NameSpace(param.Namespace).Upgrade(
		name, cht,
		helm3.WithUpgradeInstall(true),
		helm3.WithUpgradeAtoMic(false),
		helm3.WithUpgradeValues(values),
		helm3.WithUpgradeRecreatePods(true),
	)
	if err != nil {
		return nil, fmt.Errorf("upgrade or install error: %w", err)
	}
	eObj.Params = param
	eObj.Info = generateInfo(name, param)
	eObj.Version = plugin.Version

	// 持久化
	err = global.Persist.SetComponentObject(name, eObj.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return eObj, nil
}

func DeleteETCD(name string, toClean bool) (*types.ComponentETCD, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	eObj, err := cpt.TryToETCD()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}
	// 由于此组件不支持单独卸载，所以暂不实现删除secret逻辑
	// 如果需要清理存储，则需要wait等待
	_, err = global.HelmCli.NameSpace(eObj.Params.Namespace).
		Uninstall(eObj.Name, helm3.WithUninstallWait(toClean, 10*time.Minute), helm3.WithUninstallIgnoreNotFound(true))
	if err != nil {
		return nil, fmt.Errorf("uninstall component error: %w", err)
	}

	if toClean {
		err := clearETCD(eObj.Params)
		if err != nil {
			return nil, fmt.Errorf("clear component error: %w", err)
		}
	}

	err = global.Persist.DelComponent(name)
	if err != nil {
		return nil, fmt.Errorf("delete component error: %w", err)
	}

	return eObj, nil
}

func GetETCD(name string) (*types.ComponentETCD, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	eObj, err := cpt.TryToETCD()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	return eObj, nil
}

func ListETCD() ([]*types.ComponentETCD, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentETCD, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			eObj, err := cpt.TryToETCD()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, eObj)
		}
	}
	return result, nil
}
