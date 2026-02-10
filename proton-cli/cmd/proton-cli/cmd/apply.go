/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

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

// applyCmd represents the apply command
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

		// Determine which namespace to use
		nsToUse := ""
		if namespace != "" {
			// Command line namespace takes precedence
			nsToUse = namespace
		} else if conf.Deploy != nil && conf.Deploy.Namespace != "" {
			// Fall back to config file namespace
			nsToUse = conf.Deploy.Namespace
		}

		// Update the local configuration file if a namespace is specified
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
	rootCmd.AddCommand(applyCmd)
	applyCmd.PersistentFlags().StringVarP(&configPath,
		"file",
		"f",
		"",
		"proton cluster conf path")
	if err := applyCmd.MarkPersistentFlagRequired("file"); err != nil {
		panic(err)
	}

	// Use empty string as default value to allow falling back to config file
	applyCmd.PersistentFlags().StringVarP(&namespace,
		"namespace",
		"n",
		"",
		"namespace to use for deployment, overrides the namespace in config file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
