package images

import (
	"github.com/spf13/cobra"
)

var ClusterConfigFilePath string

var imageCmd = &cobra.Command{
	Use:   "images",
	Short: "manage images",
}

func SetImageCmd() *cobra.Command {
	return imageCmd
}
