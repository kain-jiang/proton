//go:build windows

package logger

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
)

type textFormatter struct {
	CallerPrettyfier func(frame *runtime.Frame) (function string, file string)
}

func NewLogger() *logrus.Logger {
	Logger := logrus.New()
	level, err := logrus.ParseLevel(global.LoggerLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)
	textFormatter := &textFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := filepath.Base(frame.File)
			funcStr := fmt.Sprintf("%s:%d", frame.Function, frame.Line)
			return funcStr, fileName
		},
	}
	Logger.SetFormatter(textFormatter)
	Logger.SetReportCaller(true)
	return Logger
}

// Format implement the Formatter interface
func (tf *textFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	// entry.Message
	funcStr, _ := tf.CallerPrettyfier(entry.Caller)
	if entry.Level == logrus.ErrorLevel {
		b.WriteString(fmt.Sprintf("%s [\u001B[1;37;41m%s\u001B[0m] func=%s  message= %s \n", entry.Time.Format("2006-01-02 15:04:05"), entry.Level, funcStr[37:], entry.Message))
	} else {
		b.WriteString(fmt.Sprintf("%s [%s] func=%s  message= %s \n", entry.Time.Format("2006-01-02 15:04:05"), entry.Level, funcStr[37:], entry.Message))
	}
	return b.Bytes(), nil
}
