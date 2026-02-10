package completion

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestCompleteFirewall(t *testing.T) {
	tests := []struct {
		name string
		conf *configuration.Firewall
		want *configuration.Firewall
	}{
		{
			name: "full",
			conf: &configuration.Firewall{Mode: configuration.FirewallFirewalld},
			want: &configuration.Firewall{Mode: configuration.FirewallFirewalld},
		},
		{
			name: "mode missing",
			conf: &configuration.Firewall{},
			want: &configuration.Firewall{Mode: configuration.FirewallFirewalld},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CompleteFirewall(tt.conf)
			assert.Equal(t, tt.want, tt.conf)
		})
	}
}
