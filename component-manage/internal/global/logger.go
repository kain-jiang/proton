package global

import (
	"strings"
	"sync"

	"component-manage/internal/config"

	"github.com/sirupsen/logrus"
)

var (
	Logger     logrus.FieldLogger
	loggerOnce sync.Once
)

func InitLogger(c *config.Config) {
	loggerOnce.Do(func() {
		logger := logrus.New()
		switch strings.ToLower(strings.TrimSpace(c.Log.Format)) {
		case "json":
			logger.SetFormatter(&logrus.JSONFormatter{})
		case "text":
			logger.SetFormatter(&logrus.TextFormatter{})
		default:
			logger.SetFormatter(&logrus.TextFormatter{})
		}
		level, err := logrus.ParseLevel(strings.ToLower(strings.TrimSpace(c.Log.Level)))
		if err != nil {
			logrus.WithError(err).Error("parse log level failed, default to debug")
			level = logrus.DebugLevel
		}
		logger.SetLevel(level)
		Logger = logger.WithFields(logrus.Fields{
			"app": "component-manage",
		})
	})
}
