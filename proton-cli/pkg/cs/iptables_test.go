package cs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateIPTablesArgsListForCleaning(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "iptables",
		},
		{
			name: "ip6tables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", tt.name))
			require.NoError(t, err, "open testdata")
			defer f.Close()

			got, err := generateIPTablesArgsListForCleaning(f)
			require.NoError(t, err)

			for _, args := range got {
				cmd := exec.Command(tt.name, args...)
				t.Log(cmd)
			}
		})
	}
}
