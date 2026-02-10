/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/push"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

var (
	username    string // Account name on repo, used for Authentication when push charts
	password    string // Account password on repo, used for Authentication when push charts
	repoUrl     string
	packagePath string
)

// pushChartsCmd represents the pushCharts command
var pushChartsCmd = &cobra.Command{
	Use:   "push-charts",
	Short: "push charts to repo by compressed archive file or directory",
	Example: `
proton-cli push-charts --package /path/to/compressedArchive.tar.gz
proton-cli push-charts --package /path/to/directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lg := logger.NewLogger()

		lg.Debugf("%#v", version.Get())

		ociPkgPath, err := filepath.Abs(packagePath)
		if err != nil {
			return fmt.Errorf("unable get absolute path of charts package: %w", err)
		}
		lg.Debugf("charts package: %s", ociPkgPath)

		chartsDir := packagePath
		if fi, err := os.Stat(packagePath); err != nil {
			return err
		} else if !fi.IsDir() {
			// Decompress the compressed archive file
			dir, err := os.MkdirTemp(os.TempDir(), "charts")
			lg.Debugf("charts dir: %s", dir)
			if err != nil {
				return err
			}
			defer os.RemoveAll(dir)
			if err = archiver.Unarchive(packagePath, dir); err != nil {
				return fmt.Errorf("Decompress %s failed: %v", packagePath, err)
			}
			chartsDir = dir
		}
		return push.PushCharts(push.ChartPushOpts{
			HelmRepo:  repoUrl,
			Username:  username,
			Password:  password,
			ChartsDir: chartsDir,
		})
	},
}

func init() {
	rootCmd.AddCommand(pushChartsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushChartsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pushChartsCmd.Flags().StringVarP(&username, "username", "u", "", "Username used in chart repo authentication")
	pushChartsCmd.Flags().StringVarP(&password, "password", "p", "", "Password used in chart repo authentication")
	pushChartsCmd.Flags().StringVar(&repoUrl, "helm-repo", "", "Repo url for push charts to. eg: https://repo.domain/chartrepo/project")
	pushChartsCmd.Flags().StringVar(&packagePath, "package", "", "Directory where charts is located or compressed archive file containing charts")
	pushChartsCmd.MarkFlagsRequiredTogether("username", "password")
	if err := pushChartsCmd.MarkFlagRequired("package"); err != nil {
		panic(err)
	}
}
