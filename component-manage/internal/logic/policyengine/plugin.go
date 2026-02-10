package policyengine

import (
	"fmt"

	"component-manage/pkg/util"

	"component-manage/internal/global"
	"component-manage/internal/pkg/cerr"
	"component-manage/pkg/models/types"
)

const (
	SelfPluginName    = "policyengine"
	SelfPluginType    = "policyengine"
	SelfComponentType = "policyengine"
)

// EnablePolicyEnginePlugin 启用 policyengine 插件
func EnablePolicyEnginePlugin(param types.PolicyEnginePluginConfig) error {
	// params 转 map
	m, err := util.ToMap(param)
	if err != nil {
		return fmt.Errorf("params to map error: %w", err)
	}

	pluginRelease := &types.PluginObject{
		Name:    SelfPluginName,
		Type:    SelfPluginType,
		Version: param.ChartVersion,
		Config:  m,
	}

	if err := global.Persist.SetPluginObject(SelfPluginName, pluginRelease); err != nil {
		return fmt.Errorf("set plugin policyengine error: %w", err)
	}
	return nil
}

// UpgradePolicyEnginePlugin 升级 policyengine 插件
func UpgradePolicyEnginePlugin(param types.PolicyEnginePluginConfig) error {
	return EnablePolicyEnginePlugin(param)
}

func GetPolicyEnginePlugin() (*types.PluginObject, error) {
	plugin, err := global.Persist.GetPluginObject(SelfPluginName)
	if err != nil {
		return nil, fmt.Errorf("get plugin error: %w", err)
	}
	if plugin == nil {
		return nil, cerr.NewError(cerr.PluginNotFoundError, "plugin policyengine not found", "")
	}
	return plugin, nil
}
