package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type FileHook struct {
	FileName string
}

func (h *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *FileHook) Fire(entry *logrus.Entry) error {
	file, err := os.OpenFile(h.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	line, err := entry.String()
	if err != nil {
		return err
	}

	_, err = file.WriteString(line)
	if err != nil {
		return err
	}

	return nil
}
