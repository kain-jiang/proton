package completion

import (
	"testing"

	"github.com/go-test/deep"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestCompletionChrony(t *testing.T) {
	tests := []struct {
		name string
		ch   *configuration.Chrony
		cs   *configuration.Cs
		want *configuration.Chrony
	}{
		{
			name: "empty",
			ch:   nil,
			cs: &configuration.Cs{
				Master: []string{
					"node-71-59",
				},
			},
			want: &configuration.Chrony{
				Mode:   configuration.ChronyModeUserManaged,
				Server: []string{},
			},
		},
		{
			name: "localmaster",
			ch: &configuration.Chrony{
				Mode:   configuration.ChronyModeLocalMaster,
				Server: []string{},
			},
			cs: &configuration.Cs{
				Master: []string{
					"node-71-59",
					"node-71-60",
				},
			},
			want: &configuration.Chrony{
				Mode: configuration.ChronyModeLocalMaster,
				Server: []string{
					"node-71-59",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CompleteChrony(tt.ch, tt.cs)
			for _, diff := range deep.Equal(c, tt.want) {
				t.Error(diff)
			}
		})
	}
}
