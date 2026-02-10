package componentmanage

import (
	"errors"
	"fmt"

	componentmanageCli "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/componentmanage"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

func (m *Resetter) Reset() error {
	if !global.ClearData {
		// 不清理数据的话无需重置
		return nil
	}
	return m.resetComponent()
}

func (m *Resetter) resetComponent() error {
	cli, err := componentmanageCli.New(m.Namespace, "component-manage", 80, logger.NewLogger(), global.ComponentManageDirectConnect)
	if err != nil {
		return fmt.Errorf("new component-manage cli error: %s", err)
	}
	err = errors.Join(
		m.resetZookeeper(cli, zkName),
		m.resetKafka(cli, kafkaName),
		m.resetOpensearch(cli, opsearchName),
		m.resetRedis(cli, redisName),
		m.resetETCD(cli, etcdName),
		m.resetPolicyEngine(cli, pName),
		m.resetMariaDB(cli, mariadbName),
		m.resetMongoDB(cli, mongodbName),
		m.resetNebula(cli, nebulaName),
	)
	if err != nil {
		return fmt.Errorf("reset component-manage error: %s", err)
	}

	return nil
}
