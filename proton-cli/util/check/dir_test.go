package check

import (
	"testing"
)

func TestNodeDirAvailableChecker_Check(t *testing.T) {
	t.Skip("unimplemented")
}

func TestNodeDirAvailableChecker_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *NodeDirAvailableChecker
		want string
	}{
		{
			name: "example",
			c:    &NodeDirAvailableChecker{Node: "node-example", Path: "/var/lib/something"},
			want: "node-example-NodeDirAvailableChecker--var-lib-something",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("NodeDirAvailableChecker.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
