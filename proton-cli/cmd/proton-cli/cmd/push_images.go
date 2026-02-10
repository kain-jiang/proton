/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/push"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

var (
	registry      string
	prePullImages bool
)

// pushImagesCmd represents the pushImages command
var pushImagesCmd = &cobra.Command{
	Use:   "push-images",
	Short: "push images to repo by archive file or directory",
	Example: `
proton-cli push-images --package /path/to/archive.tar
proton-cli push-images --package /path/to/directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lg := logger.NewLogger()

		lg.Debugf("%#v", version.Get())

		ociPkgPath, err := filepath.Abs(packagePath)
		if err != nil {
			return fmt.Errorf("unable get absolute path of oci package: %w", err)
		}

		return push.PushImages(push.ImagePushOpts{
			Registry:       registry,
			Username:       username,
			Password:       password,
			OCIPackagePath: ociPkgPath,
			PrePullImages:  true,
		}, _WorkDir)
	},
}

func init() {
	rootCmd.AddCommand(pushImagesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushImagesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pushImagesCmd.Flags().StringVarP(&username, "username", "u", "", "Username used in docker registry authentication")
	pushImagesCmd.Flags().StringVarP(&password, "password", "p", "", "Password used in docker registry authentication")
	pushImagesCmd.Flags().StringVar(&registry, "registry", "", "Registry address for push images to")
	pushImagesCmd.Flags().StringVar(&packagePath, "package", "", "Directory format as OCI or OCI tar file which contains OCI images")
	pushImagesCmd.Flags().StringVarP(&_WorkDir, "workdir", "w", "", "推送镜像使用的临时存储工作目录, 默认为系统临时目录。如unix的/tmp目录")
	pushImagesCmd.MarkFlagsRequiredTogether("username", "password")
	if err := pushImagesCmd.MarkFlagRequired("package"); err != nil {
		panic(err)
	}
}
