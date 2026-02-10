package cr

import (
	"os"
	"path/filepath"
)

const AishuContainerRegistryHostname = "acr.aishu.cn"

func DockerCLIConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".docker", "config.json")
}
