/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	rgr "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/registry"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/componentmanage"
	cmpkg "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/componentmanage/packages"
)

var componentPackage string
var componentClusterConf string
var componentTmpDir string
var componentPullImage bool

func init() {
	componentApplyCmd.Flags().StringVar(&componentPackage, "package", "", "ComponentPackage file path")
	if err := componentApplyCmd.MarkFlagRequired("package"); err != nil {
		panic(err)
	}
	componentApplyCmd.PersistentFlags().StringVarP(&componentClusterConf, "file", "f", "", "proton cluster conf path")
	componentApplyCmd.PersistentFlags().StringVar(&componentTmpDir, "tempdir", "", "component tempary directory")
	componentApplyCmd.PersistentFlags().BoolVar(&componentPullImage, "pull", true, "pull image when push images")
	componentCmd.AddCommand(componentApplyCmd)
	rootCmd.AddCommand(componentCmd)
}

var componentCmd = &cobra.Command{
	Use:     "component",
	Aliases: []string{"components"},
	Short:   "manage data component",
}

var componentApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply components by componentPackage",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var newCfg, oldCfg *configuration.ClusterConfig
		_, k := client.NewK8sClient()
		if k == nil {
			return client.ErrKubernetesClientSetNil
		}
		oldCfg, err = configuration.LoadFromKubernetes(context.Background(), k)
		if err != nil {
			return err
		}
		if componentClusterConf == "" {
			newCfg = oldCfg
		} else {
			_configPath, err := filepath.Abs(componentClusterConf)
			if err != nil {
				return fmt.Errorf("unable get absolute path of config file: %w", err)
			}

			newCfg, err = configuration.LoadFromFile(_configPath)
			if err != nil {
				return err
			}
		}

		rgrCli, err := rgr.New(rgr.ConfigForCR(newCfg.Cr))
		if err != nil {
			return fmt.Errorf("create registry client fail: %w", err)
		}

		pkgs := cmpkg.NewPackage(componentPackage, log, componentTmpDir)
		if err := pkgs.Push(newCfg, componentPullImage); err != nil {
			return fmt.Errorf("push component package fail: %w", err)
		}

		up, err := componentmanage.NewComponentApply(oldCfg, newCfg, rgrCli, pkgs)
		if err != nil {
			return err
		}

		if err := up.Apply(); err != nil {
			return err
		}
		// Save Configuration
		if err := configuration.UploadToKubernetes(context.Background(), up.NewCfg, k); err != nil {
			return fmt.Errorf("unable to upload cluster config to kubernetes: %w", err)
		}

		log.Info("apply components success")
		fmt.Printf("\033[1;37;42m%s\033[0m\n", "Apply components success")
		return nil
	},
}
