package global

import (
	"sync"

	"component-manage/internal/config"

	"github.com/sirupsen/logrus"
)

var (
	Config     *config.Config
	configOnce sync.Once
)

func InitConfig(configPath string) {
	configOnce.Do(func() {
		cfg, err := config.NewConfig(configPath)
		if err != nil {
			logrus.WithError(err).Fatal("create config failed")
		}
		Config = cfg
	})
}
