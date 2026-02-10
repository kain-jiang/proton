package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/protonk8s"
)

var kubernetesCmd = &cobra.Command{
	Use:   "kubernetes",
	Short: "manage kubernetes",
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "show kubernetes",
	Example: `
proton-cli kubernetes show`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("display kubernetes info")
		_, k := client.NewK8sClient()
		if k == nil {
			return client.ErrKubernetesClientSetNil
		}
		c := protonk8s.Config{}
		if err := c.GetCurrentCalicoConfig(k); err != nil {
			return err
		}
		log.Printf("calico version: %s\n", c.CurrentVersion)
		return nil
	},
}

var calicoCmd = &cobra.Command{
	Use:   "calico",
	Short: "manage kubernetes calico plugin",
}

var calicoUpgradeCmd = &cobra.Command{
	Use:  "upgrade [version]",
	Args: cobra.ExactArgs(1),
	Example: `
proton-cli kubernetes calico upgrade [version]`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("Upgrading Calico to version %s\n", args[0])
		// break if args[0] is empty
		if len(args[0]) == 0 {
			return fmt.Errorf("calico version: [%s] is empty.Only support %s", args[0], protonk8s.GetSupportedVersion())
		}
		if _, ok := protonk8s.SupportedVersion[args[0]]; !ok {
			return fmt.Errorf("calico version: %s not support.Only support %s", args[0], protonk8s.GetSupportedVersion())
		}
		_, k := client.NewK8sClient()
		if k == nil {
			return fmt.Errorf("failed to get dynamic kubernetes client, error: %w", client.ErrKubernetesClientSetNil)
		}
		extClient := client.NewExtK8sClient()
		if extClient == nil {
			return fmt.Errorf("failed to get extension kubernetes client, error: %w", client.ErrKubernetesExtensionClientNil)
		}
		c := protonk8s.Config{
			Version: args[0],
		}
		return c.CalicoUpgrade(extClient, k)
	},
}

func init() {
	kubernetesCmd.AddCommand(showCmd)
	kubernetesCmd.AddCommand(calicoCmd)
	calicoCmd.AddCommand(calicoUpgradeCmd)
}

func K8SCmd() *cobra.Command {
	return kubernetesCmd
}
