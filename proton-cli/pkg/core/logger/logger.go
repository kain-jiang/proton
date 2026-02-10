//go:build linux

package logger

import (
	"bytes"
	"fmt"
	"log/syslog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	lsyslog "github.com/sirupsen/logrus/hooks/syslog"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
)

type textFormatter struct {
	CallerPrettyfier func(frame *runtime.Frame) (function string, file string)
}

func NewLogger(opts ...string) *logrus.Logger {
	var useDeprecated bool
	for _, o := range opts {
		if o == "deprecated" {
			useDeprecated = true
		}
	}
	if !useDeprecated {
		return NewLoggerNG()
	}

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
	hook, err := lsyslog.NewSyslogHook("", "", syslog.LOG_INFO|syslog.LOG_DEBUG|syslog.LOG_ERR, "proton-cli")
	if err == nil {
		Logger.Hooks.Add(hook)
	}
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

// NewLoggerNG return a next generation logger.
func NewLoggerNG() *logrus.Logger {
	var prettyfier = func(f *runtime.Frame) (function, file string) {
		file = fmt.Sprintf("%16s:%-4d", filepath.Base(f.File), f.Line)
		return
	}
	var formatter = logrus.TextFormatter{
		DisableQuote:     true,
		FullTimestamp:    true,
		TimestampFormat:  "15:04:05.000",
		QuoteEmptyFields: true,
		CallerPrettyfier: prettyfier,
	}
	level, err := logrus.ParseLevel(global.LoggerLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	var logger = logrus.Logger{
		Out:          os.Stderr,
		Hooks:        make(logrus.LevelHooks),
		Formatter:    &formatter,
		ReportCaller: true,
		Level:        level,
	}
	if hook, err := lsyslog.NewSyslogHook("", "", syslog.LOG_INFO|syslog.LOG_DEBUG|syslog.LOG_ERR, "proton-cli"); err == nil {
		logger.Hooks.Add(hook)
	}

	return &logger
}
