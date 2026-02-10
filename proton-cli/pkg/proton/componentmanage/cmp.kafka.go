package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

func (m *Applier) applyKafka(cli componentmanageCli.Client, name, zkName string) error {
	c := m.charts.Get("proton-kafka", "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", "proton-kafka")
		return nil
	}

	toUpgrade, oldVersion, err := cli.ComponentUpgradable("kafka", name, c.Metadata.Version)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, c.Metadata.Version)
		return nil
	}

	err = cli.EnableKafka(c.Metadata.Name, c.Metadata.Version)
	if err != nil {
		return fmt.Errorf("enable kafka error: %s", err)
	}

	if m.NewCfg.Kafka == nil {
		return nil
	}

	oldKafkaInfo, err := cli.GetKafka(name)
	if err != nil {
		return fmt.Errorf("get kafka error: %s", err)
	}

	params := mustToMap(m.NewCfg.Kafka)
	params["namespace"] = configuration.GetProtonResourceNSFromFile()

	var info map[string]any
	if oldKafkaInfo != nil { // 更新
		info, err = cli.UpgradeKafka(name, params, zkName)
		if err != nil {
			return fmt.Errorf("upgrade kafka error: %s", err)
		}
	} else {
		info, err = cli.CreateKafka(name, params, zkName)
		if err != nil {
			return fmt.Errorf("create kafka error: %s", err)
		}
	}

	if m.NewCfg.ResourceConnectInfo.Mq != nil &&
		m.NewCfg.ResourceConnectInfo.Mq.SourceType == configuration.Internal &&
		m.NewCfg.ResourceConnectInfo.Mq.MqType == configuration.KafkaType {
		m.NewCfg.ResourceConnectInfo.Mq = mustFromMap[configuration.MqInfo](info)
	}
	log.Info("install/upgrade kafka success by component-management")

	return nil
}

func (m *Resetter) resetKafka(cli componentmanageCli.Client, name string) error {
	return cli.DeleteKafka(name)
}
