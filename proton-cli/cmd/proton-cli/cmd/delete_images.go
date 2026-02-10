package cmd

import (
	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"

	"context"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/cmd/proton-cli/cmd/utils"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

var ClusterConfigFilePath string

var deleteImages = &cobra.Command{
	Use:   "delete-images",
	Short: "Clear space for images and private images",
	Example: `
proton-cli delete-images`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		log := logger.NewLogger()

		log.Debugf("%#v", version.Get())
		var c *configuration.ClusterConfig
		switch {
		case ClusterConfigFilePath != "":
			c, err = configuration.LoadFromFile(ClusterConfigFilePath)
			if err != nil {
				return err
			}
		default:
			_, k := client.NewK8sClient()
			if k == nil {
				return client.ErrKubernetesClientSetNil
			}
			c, err = configuration.LoadFromKubernetes(context.Background(), k)
			if err != nil {
				return err
			}
		}

		if len(args) != 0 {
			nodes, err := utils.NodeListFromIPList(args)
			if err != nil {
				return err
			}
			c.Nodes = nodes
		}

		image := &cr.Image{
			Logger: logger.NewLogger(),
		}
		return image.ReleaseSpace(c)
	},
}

func init() {
	rootCmd.AddCommand(deleteImages)
	deleteImages.Flags().StringVarP(&ClusterConfigFilePath, "file", "f", ClusterConfigFilePath, "Clearing the nodes of the cluster config file")
}
