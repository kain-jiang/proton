/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cluster

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/apply"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

var configPath string
var namespace string

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply proton cluster by file name",
	Long: `apply proton cluster by file name,For Example:
    proton-cli apply -f conf.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lg := logger.NewLogger()

		info := version.Get()
		lg.WithFields(logrus.Fields{
			"gitVersion":   info.GitVersion,
			"gitCommit":    info.GitCommit,
			"gitTreeState": info.GitTreeState,
			"buildDate":    info.BuildDate,
			"goVersion":    info.GoVersion,
			"compiler":     info.Compiler,
			"platform":     info.Platform,
		}).Info("version info")

		configPath, err := filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("unable get absolute path of config file: %w", err)
		}

		conf, err := configuration.LoadFromFile(configPath)
		if err != nil {
			return err
		}
		lg.WithFields(logrus.Fields{
			"path":    configPath,
			"content": toJSON(conf),
		}).Debug("load config file")

		nsToUse := ""
		if namespace != "" {
			nsToUse = namespace
		} else if conf.Deploy != nil && conf.Deploy.Namespace != "" {
			nsToUse = conf.Deploy.Namespace
		}

		if nsToUse != "" {
			fmt.Printf("Updating local configuration with namespace: %s\n", nsToUse)
			if err := configuration.UpdateProtonCliEnvConfig(nsToUse); err != nil {
				return fmt.Errorf("unable to update proton-cli.yaml: %v", err)
			}
		}
		return apply.Apply(conf)
	},
	DisableSuggestions: false,
}

func init() {
	applyCmd.PersistentFlags().StringVarP(&configPath,
		"file",
		"f",
		"",
		"proton cluster conf path")
	if err := applyCmd.MarkPersistentFlagRequired("file"); err != nil {
		panic(err)
	}

	applyCmd.PersistentFlags().StringVarP(&namespace,
		"namespace",
		"n",
		"",
		"namespace to use for deployment, overrides the namespace in config file")
}

func NewApplyCommand() *cobra.Command {
	return applyCmd
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
