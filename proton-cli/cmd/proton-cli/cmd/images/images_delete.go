package images

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/cmd/proton-cli/cmd/utils"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

var n int

var imageDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "remove local images and private registry images",
	Example: `
proton-cli images delete`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		log := logger.NewLogger()

		log.Debugf("%#v", version.Get())
		var c *configuration.ClusterConfig
		var runtime string

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
			// different runtime using different api,currently surrport docker or containerd
			nodes, err := k.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
			if err != nil {
				log.Errorln("get k8s runtime failed, cause: ", err)
				return err
			}
			for _, node := range nodes.Items {
				if strings.Contains(node.Status.NodeInfo.ContainerRuntimeVersion, "docker") {
					log.Infoln("current k8s runtime is ", node.Status.NodeInfo.ContainerRuntimeVersion, "using docker engine api to clean")
					runtime = "docker"
				} else if strings.Contains(node.Status.NodeInfo.ContainerRuntimeVersion, "containerd") {
					log.Infoln("current k8s runtime is ", node.Status.NodeInfo.ContainerRuntimeVersion, "using containerd api to clean")
					runtime = "containerd"
				}
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
		if runtime == "containerd" {
			// TODO:
			return nil
		}
		return image.ReleaseDockerSpace(c, n)
	},
}

func init() {
	imageCmd.AddCommand(imageDeleteCmd)
	imageDeleteCmd.Flags().IntVarP(&n, "number", "n", 3, "image retain number after exec images delete")
}
