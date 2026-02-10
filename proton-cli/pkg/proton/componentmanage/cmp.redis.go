package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

func (m *Applier) applyRedis(cli componentmanageCli.Client, name string) error {
	c := m.charts.Get("proton-redis", "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", "proton-redis")
		return nil
	}

	toUpgrade, oldVersion, err := cli.ComponentUpgradable("redis", name, c.Metadata.Version)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, c.Metadata.Version)
		return nil
	}

	err = cli.EnableRedis(c.Metadata.Name, c.Metadata.Version)
	if err != nil {
		return fmt.Errorf("enable redis error: %s", err)
	}

	if m.NewCfg.Proton_redis == nil {
		return nil
	}

	oldRInfo, err := cli.GetRedis(name)
	if err != nil {
		return fmt.Errorf("get redis error: %s", err)
	}

	params := mustToMap(m.NewCfg.Proton_redis)
	params["namespace"] = configuration.GetProtonResourceNSFromFile()

	var info map[string]any
	if oldRInfo != nil {
		// 更新
		info, err = cli.UpgradeRedis(name, params)
		if err != nil {
			return fmt.Errorf("upgrade redis error: %s", err)
		}
	} else {
		info, err = cli.CreateRedis(name, params)
		if err != nil {
			return fmt.Errorf("create redis error: %s", err)
		}
	}

	m.NewCfg.ResourceConnectInfo.Redis = mustFromMap[configuration.RedisInfo](info)
	log.Info("install/upgrade redis success by component-management")

	return nil
}

func (m *Resetter) resetRedis(cli componentmanageCli.Client, name string) error {
	return cli.DeleteRedis(name)
}
