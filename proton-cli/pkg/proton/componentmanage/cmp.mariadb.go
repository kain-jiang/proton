package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

const (
	mariadbOperatorChartName = "rds-mariadb-operator"
	mariadbOperatorNamespace = "default"

	mariadbAdminAccountSecretName = "proton-mariadb-proton-rds"

	mariadbManagementServiceName = "mariadb-mgmt-cluster"
	mariadbManagementServicePort = 8888

	mariadbEtcdImageRepository = "proton/etcd"
	mariadbEtcdImageTag        = "v3.3.19"

	mariadbExporterImageRepository = "proton/rds-exporter"
	mariadbExporterImageTag        = "2.1.0"

	mariadbMgmtImageRepository = "proton/rds-mgmt"
	mariadbMgmtImageTag        = "2.3.0"

	mariadbMariaDBImageRepository = "proton/rds-mariadb"
	mariadbMariaDBImageTag        = "2.0.6"
)

func (m *Applier) applyMariaDB(cli componentmanageCli.Client, name string) error {
	if m.NewCfg.Deploy.Namespace != "" {
		log.Info("custom namespace is enabled, skip apply mariadb")
		return nil
	}

	c := m.charts.Get(mariadbOperatorChartName, "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", mariadbOperatorChartName)
		return nil
	}

	newVersion := fmt.Sprintf("%s+%s", c.Metadata.Version, m.SearchTag(originRegistry, mariadbMariaDBImageRepository, mariadbMariaDBImageTag))
	toUpgrade, oldVersion, err := cli.ComponentUpgradable("mariadb", name, newVersion)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, newVersion)
		return nil
	}

	err = cli.EnableMariaDB(componentmanageCli.MariaDBPluginInfo{
		ChartName:    c.Metadata.Name,
		ChartVersion: c.Metadata.Version,
		Namespace:    mariadbOperatorNamespace,
		Images: componentmanageCli.MariaDBPluginImagesInfo{
			MariaDB:  m.SearchImage(originRegistry, mariadbMariaDBImageRepository, mariadbMariaDBImageTag),
			ETCD:     m.SearchImage(originRegistry, mariadbEtcdImageRepository, mariadbEtcdImageTag),
			Exporter: m.SearchImage(originRegistry, mariadbExporterImageRepository, mariadbExporterImageTag),
			Mgmt:     m.SearchImage(originRegistry, mariadbMgmtImageRepository, mariadbMgmtImageTag),
		},
	})
	if err != nil {
		return fmt.Errorf("enable mariadb error: %s", err)
	}

	if m.NewCfg.Proton_mariadb == nil {
		return nil
	}

	mariadbInfo, err := cli.GetMariaDB(name)
	if err != nil {
		return fmt.Errorf("get mariadb error: %s", err)
	}

	params := mustToMap(m.NewCfg.Proton_mariadb)
	//额外参数，使用内置数据库时连接信息必须包含业务账户用户名密码
	params["username"] = m.NewCfg.ResourceConnectInfo.Rds.Username
	params["password"] = m.NewCfg.ResourceConnectInfo.Rds.Password
	params["namespace"] = configuration.GetProtonResourceNSFromFile()
	params["admin_secret_name"] = mariadbAdminAccountSecretName

	var info map[string]any
	if mariadbInfo != nil {
		// 更新
		info, err = cli.UpgradeMariaDB(name, params)
		if err != nil {
			return fmt.Errorf("upgrade mariadb error: %s", err)
		}
	} else {
		info, err = cli.CreateMariaDB(name, params)
		if err != nil {
			return fmt.Errorf("create mariadb error: %s", err)
		}
	}
	if m.NewCfg.ResourceConnectInfo.Rds.SourceType != configuration.External {
		// 是外置连接信息是不进行更新
		m.NewCfg.ResourceConnectInfo.Rds = mustFromMap[configuration.RdsInfo](info)
	}
	log.Info("install/upgrade mariadb success by component-management")

	return nil
}

func (m *Resetter) resetMariaDB(cli componentmanageCli.Client, name string) error {
	return cli.DeleteMariaDB(name)
}
