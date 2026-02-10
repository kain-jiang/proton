package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

func (m *Applier) applyETCD(cli componentmanageCli.Client, name string) error {
	c := m.charts.Get("proton-etcd", "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", "proton-etcd")
		return nil
	}

	toUpgrade, oldVersion, err := cli.ComponentUpgradable("etcd", name, c.Metadata.Version)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, c.Metadata.Version)
		return nil
	}

	err = cli.EnableETCD(c.Metadata.Name, c.Metadata.Version)
	if err != nil {
		return fmt.Errorf("enable etcd error: %s", err)
	}

	if m.NewCfg.Proton_etcd == nil {
		return nil
	}

	oldEInfo, err := cli.GetETCD(name)
	if err != nil {
		return fmt.Errorf("get etcd error: %s", err)
	}

	params := mustToMap(m.NewCfg.Proton_etcd)
	params["namespace"] = configuration.GetProtonResourceNSFromFile()

	var info map[string]any
	if oldEInfo != nil {
		// 更新
		info, err = cli.UpgradeETCD(name, params)
		if err != nil {
			return fmt.Errorf("upgrade etcd error: %s", err)
		}
	} else {
		info, err = cli.CreateETCD(name, params)
		if err != nil {
			return fmt.Errorf("create etcd error: %s", err)
		}
	}

	m.NewCfg.ResourceConnectInfo.Etcd = mustFromMap[configuration.EtcdInfo](info)
	log.Info("install/upgrade etcd success by component-management")

	return nil
}

func (m *Resetter) resetETCD(cli componentmanageCli.Client, name string) error {
	return cli.DeleteETCD(name)
}
