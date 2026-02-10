package mariadb

import (
	"fmt"
	"reflect"
	"time"

	"component-manage/internal/logic/base"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/k8s"
	"component-manage/pkg/models/types"
	"component-manage/pkg/util"
)

func CreateMariaDB(name string, param *types.MariaDBComponentParams) (*types.ComponentMariaDB, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin mariadb not found", "")
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
	if err := checkMariaDB(name, param, nil); err != nil {
		// return nil, fmt.Errorf("check zookeeper params error: %w", err)
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check mariadb params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.MariaDBPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	crManifest := mariadbManifest(name, param, plgInfo)
	if err := prepareMariaDB(param); err != nil {
		return nil, fmt.Errorf("prepare mariadb error: %w", err)
	}

	global.Logger.WithField("manifest", crManifest).Debug("display mariadb manifest")

	// 创建CR
	_, err = global.K8sCli.CustomResourceSet(k8s.GVRMariaDB, name, param.Namespace, crManifest)
	if err = base.DealK8sStatusError(err); err != nil {
		return nil, fmt.Errorf("set custom resource failed: %w", err)
	}

	componentRelease := &types.ComponentMariaDB{
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

func UpgradeMariaDB(name string, param *types.MariaDBComponentParams) (*types.ComponentMariaDB, error) {
	// 查询插件是否存在
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin mariadb not found", "")
	}

	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	obj, err := cpt.TryToMariaDB()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	// 检查参数（按需，将完善参数）
	if err := checkMariaDB(name, param, obj.Params); err != nil {
		return nil, cerr.NewError(cerr.ParamsInvalidError, "check mariadb params error", err.Error())
	}

	// 插件信息
	plgInfo, err := util.FromMap[types.MariaDBPluginConfig](plugin.Config)
	if err != nil {
		return nil, fmt.Errorf("parse plugin error: %w", err)
	}

	crManifest := mariadbManifest(name, param, plgInfo)
	if err := prepareMariaDB(param); err != nil {
		return nil, fmt.Errorf("prepare mariadb error: %w", err)
	}

	global.Logger.WithField("manifest", crManifest).Debug("display mariadb manifest")

	// 创建/更新CR
	_, err = global.K8sCli.CustomResourceSet(k8s.GVRMariaDB, name, param.Namespace, crManifest)
	if err = base.DealK8sStatusError(err); err != nil {
		return nil, fmt.Errorf("set custom resource failed: %w", err)
	}

	obj.Params = param
	newInfo := generateInfo(name, param)

	// 如何判断两个结构体字段是否一致？
	if !reflect.DeepEqual(obj.Info, newInfo) {
		// 发生了变更，依赖需要更新
		global.Logger.WithField("component", name).WithField("info", newInfo).Warn("component info is changed")
	}

	obj.Info = newInfo
	obj.Version = plugin.Version
	// 持久化
	err = global.Persist.SetComponentObject(name, obj.ToBase())
	if err != nil {
		return nil, fmt.Errorf("persist component error: %w", err)
	}

	return obj, nil
}

func DeleteMariaDB(name string, toClean bool) (*types.ComponentMariaDB, error) {
	// 检查组件已存在
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	mariadbObj, err := cpt.TryToMariaDB()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}

	err = global.K8sCli.CustomResourceDelete(k8s.GVRMariaDB, mariadbObj.Name, mariadbObj.Params.Namespace)
	if err != nil {
		return nil, fmt.Errorf("delete custom resource failed: %w", err)
	}

	if toClean {
		// 无法立刻删除， 这里暂时先sleep 30s
		time.Sleep(30 * time.Second)
		err := clearMariaDB(mariadbObj.Params)
		if err != nil {
			return nil, fmt.Errorf("clear component error: %w", err)
		}
	}

	err = global.Persist.DelComponent(name)
	if err != nil {
		return nil, fmt.Errorf("delete component error: %w", err)
	}

	return mariadbObj, nil
}

func GetMariaDB(name string) (*types.ComponentMariaDB, error) {
	cpt, err := global.Persist.GetComponentObject(name)
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}
	if cpt == nil {
		return nil, cerr.NewError(cerr.ComponentNotFoundError, "component not found", "")
	}
	obj, err := cpt.TryToMariaDB()
	if err != nil {
		return nil, fmt.Errorf("parse component error: %w", err)
	}
	return obj, nil
}

func ListMariaDB() ([]*types.ComponentMariaDB, error) {
	// 检查组件已存在
	cpts, err := global.Persist.GetAllComponentObject()
	if err != nil {
		return nil, fmt.Errorf("get component error: %w", err)
	}

	result := make([]*types.ComponentMariaDB, 0)

	for _, cpt := range cpts {
		if cpt.Type == SelfComponentType {
			obj, err := cpt.TryToMariaDB()
			if err != nil {
				return nil, fmt.Errorf("parse component error: %w", err)
			}
			result = append(result, obj)
		}
	}
	return result, nil
}
