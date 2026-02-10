package cr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDockerCLIConfigPath(t *testing.T) {
	h, err := os.UserHomeDir()
	if err != nil {
		t.Error(err)
	}

	if got, want := DockerCLIConfigPath(), filepath.Join(h, ".docker", "config.json"); got != want {
		t.Errorf("DockerCLIConfigPath() = %v, want %v", got, want)
	}
}
