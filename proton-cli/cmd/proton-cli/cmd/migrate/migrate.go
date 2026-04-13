/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/migrate"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

// templateCmd represents the template command
var configPath string
var migrateMode string
var tlsSecretName string
var certificatePath string
var keyPath string
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate components deployed by other programs.",
	Long: `migrate components deployed by other programs so that they can be managed by the proton-cli. For example:
    proton-cli migrate eceph-and-anyshare
Currently only ECeph deployed with anyshare is supported by this command.`,
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
		pdCfg := new(configuration.ProtonDeployClusterConfig)
		if migrateMode == global.MigrateECephAndAnyShare {
			configPath, err := filepath.Abs(configPath)
			if err != nil {
				return fmt.Errorf("unable get absolute path of config file: %w", err)
			}
			bytes, err := os.ReadFile(configPath)
			if err != nil {
				return err
			}
			if err := yaml.Unmarshal(bytes, pdCfg); err != nil {
				return err
			}
			lg.Debugf("proton-deploy config: %v", pdCfg)
		}
		//validate migrateMode
		flagMigrationModeIsSupported := false
		for _, modeItem := range global.SupportedECephMigrationMode {
			if migrateMode == modeItem {
				flagMigrationModeIsSupported = true
			}
		}
		if !flagMigrationModeIsSupported {
			return fmt.Errorf("unsupported migration mode")
		}
		return migrate.Migrate(*pdCfg, migrateMode, tlsSecretName, certificatePath, keyPath)
	},
}

func init() {
	migrateCmd.PersistentFlags().StringVarP(&configPath,
		"file",
		"f",
		"/etc/cluster.yaml",
		"proton-deploy cluster conf path")
	migrateCmd.PersistentFlags().StringVarP(&tlsSecretName,
		"secret-name",
		"",
		"",
		"name of the secret for TLS certificate storage of existing ECeph installation")
	migrateCmd.PersistentFlags().StringVarP(&certificatePath,
		"certificate-data",
		"",
		"",
		"path to the ECeph TLS certificate file to be used")
	migrateCmd.PersistentFlags().StringVarP(&keyPath,
		"key-data",
		"",
		"",
		"path to the ECeph TLS certificate key file to be used")
	migrateCmd.PersistentFlags().StringVarP(&migrateMode,
		"migrate-mode",
		"m",
		"",
		"the component to migrate")
	if err := migrateCmd.MarkPersistentFlagRequired("migrate-mode"); err != nil {
		panic(err)
	}
}

func NewMigrateCommand() *cobra.Command {
	return migrateCmd
}
