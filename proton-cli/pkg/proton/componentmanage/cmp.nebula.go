package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

const (
	nebulaOperatorChartName = "nebula-operator"

	nebulaAdminAccountSecretName = "nebula"

	nebulaTagGraphdImage   = "3.5.0"
	nebulaTagMetadImage    = "3.5.0"
	nebulaTagStoragedImage = "3.5.0"
	nebulaTagExporterImage = "3.5.0"
)

func (m *Applier) applyNebula(cli componentmanageCli.Client, name string) error {
	if m.NewCfg.Deploy.Namespace != "" {
		log.Info("custom namespace is enabled, skip apply nebula")
		return nil
	}
	c := m.charts.Get(nebulaOperatorChartName, "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", nebulaOperatorChartName)
		return nil
	}

	newVersion := fmt.Sprintf("%s+%s", c.Metadata.Version, m.SearchTag(originRegistry, "proton/nebula-graphd", nebulaTagGraphdImage))
	toUpgrade, oldVersion, err := cli.ComponentUpgradable("nebula", name, newVersion)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, newVersion)
		return nil
	}

	err = cli.EnableNebula(componentmanageCli.NebulaPluginInfo{
		ChartName:    c.Metadata.Name,
		ChartVersion: c.Metadata.Version,
		Namespace:    configuration.GetProtonResourceNSFromFile(),
		Images: componentmanageCli.NebulaPluginImagesInfo{
			GraphD:   m.SearchImage(originRegistry, "proton/nebula-graphd", nebulaTagGraphdImage),
			MetaD:    m.SearchImage(originRegistry, "proton/nebula-metad", nebulaTagMetadImage),
			StorageD: m.SearchImage(originRegistry, "proton/nebula-storaged", nebulaTagStoragedImage),
			Exporter: m.SearchImage(originRegistry, "proton/nebula-stats-exporter", nebulaTagExporterImage),
		},
	})
	if err != nil {
		return fmt.Errorf("enable nebula error: %s", err)
	}

	if m.NewCfg.Nebula == nil {
		return nil
	}

	nebulaInfo, err := cli.GetNebula(name)
	if err != nil {
		return fmt.Errorf("get nebula error: %s", err)
	}

	params := mustToMap(m.NewCfg.Nebula)

	params["admin_secret_name"] = nebulaAdminAccountSecretName
	params["namespace"] = configuration.GetProtonResourceNSFromFile()

	var relParam map[string]any

	if nebulaInfo != nil {
		// 更新
		relParam, _, err = cli.UpgradeNebula(name, params)
		if err != nil {
			return fmt.Errorf("upgrade nebula error: %s", err)
		}
	} else {
		relParam, _, err = cli.CreateNebula(name, params)
		if err != nil {
			return fmt.Errorf("create nebula error: %s", err)
		}
	}

	m.NewCfg.Nebula = mustFromMap[configuration.Nebula](relParam)

	log.Info("install/upgrade nebula success by component-management")
	return nil
}

func (m *Resetter) resetNebula(cli componentmanageCli.Client, name string) error {
	return cli.DeleteNebula(name)
}
