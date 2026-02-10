package componentmanage

import (
	"fmt"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
)

const (
	mongodbOperatorChartName = "mongodb-operator"

	mongodbAdminAccountSecretName = "mongodb-secret"

	mongodbTagMongodbImage   = "2.0.0"
	mongodbTagMgmtImage      = "2.2.4"
	mongodbTagExporterImage  = "2.2.2"
	mongodbTagLogrotateImage = "1.0.1"
)

func (m *Applier) applyMongoDB(cli componentmanageCli.Client, name string) error {
	if m.NewCfg.Deploy.Namespace != "" {
		log.Info("custom namespace is enabled, skip apply mongodb")
		return nil
	}

	c := m.charts.Get(mongodbOperatorChartName, "")
	if c == nil {
		log.Infof("chart %s not found, skip apply", mongodbOperatorChartName)
		return nil
	}

	newVersion := fmt.Sprintf("%s+%s", c.Metadata.Version, m.SearchTag(originRegistry, "proton/mongodb", mongodbTagMongodbImage))
	toUpgrade, oldVersion, err := cli.ComponentUpgradable("mongodb", name, newVersion)
	if err != nil {
		// todo
		return fmt.Errorf("check component upgradable error: %s", err)
	}
	if !toUpgrade {
		log.Infof("component %s is up to date: skip %s -> %s", name, oldVersion, newVersion)
		return nil
	}

	err = cli.EnableMongoDB(componentmanageCli.MongoDBPluginInfo{
		ChartName:    c.Metadata.Name,
		ChartVersion: c.Metadata.Version,
		Namespace:    configuration.GetProtonResourceNSFromFile(),
		Images: componentmanageCli.MongoDBPluginImagesInfo{
			MongoDB:   m.SearchImage(originRegistry, "proton/mongodb", mongodbTagMongodbImage),
			Logrotate: m.SearchImage(originRegistry, "proton/logrotate", mongodbTagLogrotateImage),
			Exporter:  m.SearchImage(originRegistry, "proton/mongodb-exporter", mongodbTagExporterImage),
			Mgmt:      m.SearchImage(originRegistry, "proton/proton-mongodb-mgmt", mongodbTagMgmtImage),
		},
	})
	if err != nil {
		return fmt.Errorf("enable mongodb error: %s", err)
	}

	if m.NewCfg.Proton_mongodb == nil {
		return nil
	}

	mongodbInfo, err := cli.GetMongoDB(name)
	if err != nil {
		return fmt.Errorf("get mongodb error: %s", err)
	}

	params := mustToMap(m.NewCfg.Proton_mongodb)
	//额外参数，使用内置数据库时连接信息必须包含业务账户用户名密码
	params["username"] = m.NewCfg.ResourceConnectInfo.Mongodb.Username
	params["password"] = m.NewCfg.ResourceConnectInfo.Mongodb.Password
	params["namespace"] = configuration.GetProtonResourceNSFromFile()
	params["admin_secret_name"] = mongodbAdminAccountSecretName

	var info map[string]any
	if mongodbInfo != nil {
		// 更新
		info, err = cli.UpgradeMongoDB(name, params)
		if err != nil {
			return fmt.Errorf("upgrade mongodb error: %s", err)
		}
	} else {
		info, err = cli.CreateMongoDB(name, params)
		if err != nil {
			return fmt.Errorf("create mongodb error: %s", err)
		}
	}

	m.NewCfg.ResourceConnectInfo.Mongodb = mustFromMap[configuration.MongodbInfo](info)
	log.Info("install/upgrade mongodb success by component-management")

	return nil
}

func (m *Resetter) resetMongoDB(cli componentmanageCli.Client, name string) error {
	return cli.DeleteMongoDB(name)
}
