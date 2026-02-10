package workspace

import "path/filepath"

func GenerateWorkspacePath(output, version, architecture string) string {
	return filepath.Join(output, version, architecture)
}
