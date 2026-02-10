package global

import (
	"sync"

	"component-manage/internal/config"
	"component-manage/internal/pkg/persist"
	"component-manage/pkg/k8s"

	"github.com/sirupsen/logrus"
)

var (
	Persist     persist.Persist
	persistOnce sync.Once
)

func InitPersist(config *config.Config, k8sCli k8s.Client, logger logrus.FieldLogger) {
	persistOnce.Do(func() {
		Persist = persist.NewK8sPersist(
			k8sCli,
			config.Persist.SecretNamespace,
			config.Persist.SecretComponentsName,
			config.Persist.SecretPluginsName,
			logger.WithField("persist", "k8s"),
		)
	})
}
