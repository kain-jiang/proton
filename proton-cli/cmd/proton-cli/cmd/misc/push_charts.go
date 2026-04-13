/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package misc

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
	chartsUsername    string
	chartsPassword    string
	chartsRepoUrl     string
	chartsPackagePath string
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

		ociPkgPath, err := filepath.Abs(chartsPackagePath)
		if err != nil {
			return fmt.Errorf("unable get absolute path of charts package: %w", err)
		}
		lg.Debugf("charts package: %s", ociPkgPath)

		chartsDir := chartsPackagePath
		if fi, err := os.Stat(chartsPackagePath); err != nil {
			return err
		} else if !fi.IsDir() {
			// Decompress the compressed archive file
			dir, err := os.MkdirTemp(os.TempDir(), "charts")
			lg.Debugf("charts dir: %s", dir)
			if err != nil {
				return err
			}
			defer os.RemoveAll(dir)
			if err = archiver.Unarchive(chartsPackagePath, dir); err != nil {
				return fmt.Errorf("Decompress %s failed: %v", chartsPackagePath, err)
			}
			chartsDir = dir
		}
		return push.PushCharts(push.ChartPushOpts{
			HelmRepo:  chartsRepoUrl,
			Username:  chartsUsername,
			Password:  chartsPassword,
			ChartsDir: chartsDir,
		})
	},
}

func init() {
	pushChartsCmd.Flags().StringVarP(&chartsUsername, "username", "u", "", "Username used in chart repo authentication")
	pushChartsCmd.Flags().StringVarP(&chartsPassword, "password", "p", "", "Password used in chart repo authentication")
	pushChartsCmd.Flags().StringVar(&chartsRepoUrl, "helm-repo", "", "Repo url for push charts to. eg: https://repo.domain/chartrepo/project")
	pushChartsCmd.Flags().StringVar(&chartsPackagePath, "package", "", "Directory where charts is located or compressed archive file containing charts")
	pushChartsCmd.MarkFlagsRequiredTogether("username", "password")
	if err := pushChartsCmd.MarkFlagRequired("package"); err != nil {
		panic(err)
	}
}

func NewPushChartsCommand() *cobra.Command {
	return pushChartsCmd
}
