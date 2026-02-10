package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

func (m *Applier) applyZookeeper(cli componentmanageCli.Client, name string) error {
	c := m.charts.Get("proton-zookeeper", "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", "proton-zookeeper")
		return nil
	}

	toUpgrade, oldVersion, err := cli.ComponentUpgradable("zookeeper", name, c.Metadata.Version)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, c.Metadata.Version)
		return nil
	}

	err = cli.EnableZookeeper(c.Metadata.Name, c.Metadata.Version)
	if err != nil {
		return fmt.Errorf("enable zookeeper error: %s", err)
	}

	if m.NewCfg.ZooKeeper == nil {
		return nil
	}

	oldZkInfo, err := cli.GetZookeeper(name)
	if err != nil {
		return fmt.Errorf("get zookeeper error: %s", err)
	}

	params := mustToMap(m.NewCfg.ZooKeeper)
	params["namespace"] = configuration.GetProtonResourceNSFromFile()

	var info map[string]any
	if oldZkInfo != nil {
		// 更新
		info, err = cli.UpgradeZookeeper(name, params)
		if err != nil {
			return fmt.Errorf("upgrade zookeeper error: %s", err)
		}
	} else {
		info, err = cli.CreateZookeeper(name, params)
		if err != nil {
			return fmt.Errorf("create kafka error: %s", err)
		}
	}

	_ = info
	log.Info("install/upgrade zookeeper success by component-management")

	return nil
}

func (m *Resetter) resetZookeeper(cli componentmanageCli.Client, name string) error {
	return cli.DeleteZookeeper(name)
}
