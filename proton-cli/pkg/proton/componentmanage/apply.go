package componentmanage

import (
	"errors"
	"fmt"
	"time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

var log = logger.NewLogger()

const (
	originRegistry = "acr.aishu.cn"

	kafkaName    = "kafka"
	zkName       = "zookeeper"
	opsearchName = "opensearch"
	redisName    = "proton-redis"
	etcdName     = "proton-etcd"
	pName        = "proton-policy-engine"
	mariadbName  = "mariadb"
	mongodbName  = "mongodb"
	nebulaName   = "nebula"
)

func (m *Applier) Apply() error {
	if !m.onlyInitComponent {
		err := m.Helm3.Upgrade(
			m.Release,
			helm3.ChartRefFromFile(m.ChartFile),
			helm3.WithUpgradeInstall(true),
			helm3.WithUpgradeAtoMic(false),
			helm3.WithUpgradeValues(m.Values),
			helm3.WithUpgradeWait(true, 10*time.Minute),
		)
		log.Info("install/upgrade component-management success")
		if err != nil {
			return fmt.Errorf("upgrade or install component-manage error: %s", err)
		}
		log.Info("init component-management start")
	}
	return m.initComponent()
}

func (m *Applier) initComponent() error {

	cli, err := componentmanageCli.New(m.Namespace, "component-manage", 80, logger.NewLogger(), global.ComponentManageDirectConnect)
	if err != nil {
		return fmt.Errorf("new component-manage cli error: %s", err)
	}

	err = errors.Join(
		m.applyZookeeper(cli, zkName),
		m.applyKafka(cli, kafkaName, zkName),
		m.applyOpensearch(cli, opsearchName),
		m.applyRedis(cli, redisName),
		m.applyETCD(cli, etcdName),
		m.applyPolicyEngine(cli, pName, etcdName),
		m.applyMariaDB(cli, mariadbName),
		m.applyMongoDB(cli, mongodbName),
		m.applyNebula(cli, nebulaName),
	)
	if err != nil {
		return fmt.Errorf("apply component-manage error: %s", err)
	}

	return nil
}
