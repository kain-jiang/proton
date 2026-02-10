package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

func (m *Applier) applyPolicyEngine(cli componentmanageCli.Client, name, eName string) error {
	c := m.charts.Get("proton-policy-engine", "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", "proton-policy-engine")
		return nil
	}

	toUpgrade, oldVersion, err := cli.ComponentUpgradable("policyengine", name, c.Metadata.Version)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, c.Metadata.Version)
		return nil
	}

	err = cli.EnablePolicyEngine(c.Metadata.Name, c.Metadata.Version)
	if err != nil {
		return fmt.Errorf("enable policyengine error: %s", err)
	}

	if m.NewCfg.Proton_policy_engine == nil {
		return nil
	}

	oldPolicyEngineInfo, err := cli.GetPolicyEngine(name)
	if err != nil {
		return fmt.Errorf("get policyengine error: %s", err)
	}

	params := mustToMap(m.NewCfg.Proton_policy_engine)
	params["namespace"] = configuration.GetProtonResourceNSFromFile()

	var info map[string]any
	if oldPolicyEngineInfo != nil { // 更新
		info, err = cli.UpgradePolicyEngine(name, params, eName)
		if err != nil {
			return fmt.Errorf("upgrade policyengine error: %s", err)
		}
	} else {
		info, err = cli.CreatePolicyEngine(name, params, eName)
		if err != nil {
			return fmt.Errorf("create policyengine error: %s", err)
		}
	}

	m.NewCfg.ResourceConnectInfo.PolicyEngine = mustFromMap[configuration.PolicyEngineInfo](info)

	log.Info("install/upgrade policyengine success by component-management")

	return nil
}

func (m *Resetter) resetPolicyEngine(cli componentmanageCli.Client, name string) error {
	return cli.DeletePolicyEngine(name)
}
