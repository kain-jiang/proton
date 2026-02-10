package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

func (m *Applier) applyOpensearch(cli componentmanageCli.Client, name string) error {
	c := m.charts.Get("proton-opensearch", "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", "proton-opensearch")
		return nil
	}

	toUpgrade, oldVersion, err := cli.ComponentUpgradable("opensearch", name, c.Metadata.Version)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, c.Metadata.Version)
		return nil
	}

	err = cli.EnableOpensearch(c.Metadata.Name, c.Metadata.Version)
	if err != nil {
		return fmt.Errorf("enable opensearch error: %s", err)
	}

	if m.NewCfg.OpenSearch == nil {
		return nil
	}

	oldOInfo, err := cli.GetOpensearch(name)
	if err != nil {
		return fmt.Errorf("get opensearch error: %s", err)
	}

	params := mustToMap(m.NewCfg.OpenSearch)
	params["namespace"] = configuration.GetProtonResourceNSFromFile()

	var info map[string]any
	if oldOInfo != nil {
		// 更新
		info, err = cli.UpgradeOpensearch(name, params)
		if err != nil {
			return fmt.Errorf("upgrade opensearch error: %s", err)
		}
	} else {
		info, err = cli.CreateOpensearch(name, params)
		if err != nil {
			return fmt.Errorf("create opensearch error: %s", err)
		}
	}

	m.NewCfg.ResourceConnectInfo.OpenSearch = mustFromMap[configuration.OpensearchInfo](info)
	log.Info("install/upgrade opensearch success by component-management")

	return nil
}

func (m *Resetter) resetOpensearch(cli componentmanageCli.Client, name string) error {
	return cli.DeleteOpensearch(name)
}
