package global

import (
	"sync"

	"component-manage/internal/config"
	"component-manage/pkg/helm3"

	"github.com/sirupsen/logrus"
)

var (
	HelmCli  helm3.Client
	helmOnce sync.Once
)

func InitHelmCli(config *config.Config, logger logrus.FieldLogger) {
	helmOnce.Do(func() {
		cli, err := helm3.NewCli(
			"default",
			logger.WithField("cli", "helm3"),
		)
		if err != nil {
			logger.WithError(err).Fatal("create helm cli failed")
		}
		HelmCli = cli
	})
}
