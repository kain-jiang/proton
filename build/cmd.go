package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var groupTargets = &cobra.Group{
	ID:    "targets",
	Title: "Targets",
}

func NewBuildCommand() *cobra.Command {
	var logLevel = slog.LevelInfo

	cmd := &cobra.Command{
		Use:           "build",
		Short:         "proton build tool",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// 设置日志级别
			slog.Debug("Set log level", "level", logLevel)
			slog.SetLogLoggerLevel(logLevel)
		},
	}

	cmd.AddGroup(groupTargets)

	// 添加构建产物
	cmd.AddCommand(newCommandProtonPackage())

	// 根据 Git 和 Azure Pipeline 生成版本号
	cmd.AddCommand(newCommandGenerateVersion())

	// 全局参数
	cmd.PersistentFlags().TextVar(&logLevel, "log-level", logLevel, fmt.Sprintf("Log level: %s, %s, %s, %s", slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError))

	return cmd
}
