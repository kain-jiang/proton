/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration/completion"
)

var getValueFlag bool
var getNamespaceFlag string

// confCmd represents the conf command
var confCmd = &cobra.Command{
	Use:   "conf",
	Short: "show proton cluster conf",
	Long: `show proton cluster conf to stdout.For Example:
    proton-cli get conf
also,can show all conf,For Example:
    proton-cli get conf -v
You can specify a namespace with -n flag, For Example:
    proton-cli get conf -n proton`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, k := client.NewK8sClient()
		if k == nil {
			return client.ErrKubernetesClientSetNil
		}

		// If namespace is specified via command line, use it to load the configuration
		c, err := configuration.LoadFromKubernetes(context.Background(), k, getNamespaceFlag)
		if err != nil {
			return err
		}

		/// 新版本升级逻辑 start
		// 新版本proton自动从组件对应的secret中补全连接信息
		if c.ResourceConnectInfo == nil {
			if err := completion.CompleteOldClusterConfFromSecret(c, k); err != nil {
				return err
			}
		}
		// 新版本升级尽量加上component-manage
		if c.ResourceConnectInfo != nil {
			c.ComponentManage = &configuration.ComponentManagement{}
		}
		if c.Deploy == nil {
			c.Deploy, err = completion.GuessDeployConfig(cmd.Context(), k, getNamespaceFlag)
			if err != nil {
				c.Deploy = &configuration.Deploy{
					Mode:       "standard",
					Devicespec: "AS10000",
				}
			}
		}
		completion.CompletionCR(c.Cr)
		/// 新版本升级逻辑 end

		b, err := yaml.Marshal(c)
		if err != nil {
			return err
		}

		if _, err := os.Stdout.Write(b); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	getCmd.AddCommand(confCmd)
	confCmd.Flags().BoolVarP(&getValueFlag,
		"value",
		"v",
		false,
		"show all proton cluster conf")

	// Add namespace flag to get command
	// Use the namespace from the config file as the default value
	defaultNamespace := configuration.GetProtonCliConfigNSFromFile()
	getCmd.PersistentFlags().StringVarP(&getNamespaceFlag,
		"namespace",
		"n",
		defaultNamespace,
		"namespace where the application is deployed")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// confCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// confCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
