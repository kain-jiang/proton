// Copyright 2021 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcd

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"go.etcd.io/etcd/client/pkg/v3/logutil"
	"go.etcd.io/etcd/pkg/v3/cobrautl"
)

const (
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

func GetLogger() *zap.Logger {
	config := logutil.DefaultZapLoggerConfig
	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	lg, err := config.Build()
	if err != nil {
		cobrautl.ExitWithError(cobrautl.ExitBadArgs, err)
	}
	return lg
}

func GetLoggerFileAndConsole(filepath string) *zap.Logger {
	opts := []zapcore.WriteSyncer{
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath,
			MaxSize:    1024 * 1024 * 1024, // megabytes
			MaxBackups: 1,
			MaxAge:     100,   //days
			Compress:   false, // disabled by default
		}),
	}

	opts = append(opts, zapcore.AddSync(os.Stdout))

	syncWriter := zapcore.NewMultiWriteSyncer(opts...)

	// 自定义时间输出格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
	}
	// 自定义日志级别显示
	customLevelEncoder := func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + level.CapitalString() + "]")
	}

	// 自定义文件：行号输出项
	customCallerEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}

	encoderConf := zapcore.EncoderConfig{
		CallerKey:      "caller_line", // 打印文件名和行数
		LevelKey:       "level_name",
		MessageKey:     "msg",
		TimeKey:        "ts",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    customLevelEncoder,  // 小写编码器
		EncodeTime:     customTimeEncoder,   // 自定义时间格式
		EncodeCaller:   customCallerEncoder, // 全路径编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// level大写染色编码器
	encoderConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConf), syncWriter, zap.NewAtomicLevelAt(zapcore.DebugLevel))
	lg := zap.New(core).WithOptions(zap.AddCaller())
	return lg
}
func GetLoggerFile(filepath string) *zap.Logger {
	var allCore []zapcore.Core
	// 打印所有级别的日志
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.DebugLevel
	})
	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	consoleDebugging := zapcore.Lock(os.Stdout)

	// for human operators.
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	hook := lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    1024 * 1024 * 1024, // megabytes
		MaxBackups: 1,
		MaxAge:     100,   //days
		Compress:   false, // disabled by default
	}
	fileWriter := zapcore.AddSync(&hook)
	allCore = append(allCore, zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority), zapcore.NewCore(consoleEncoder, fileWriter, lowPriority))
	core := zapcore.NewTee(allCore...)
	lg := zap.New(core).WithOptions(zap.AddCaller())
	return lg
}
