package firewall

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_convert4In6To4(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		ip   netip.Addr
		want netip.Addr
	}{
		{
			name: "4",
			ip:   netip.MustParseAddr("192.168.0.1"),
			want: netip.MustParseAddr("192.168.0.1"),
		},
		{
			name: "6",
			ip:   netip.MustParseAddr("fe80::250:56ff:feb4:b3fc"),
			want: netip.MustParseAddr("fe80::250:56ff:feb4:b3fc"),
		},
		{
			name: "4in6",
			ip:   netip.MustParseAddr("::ffff:10.4.71.191"),
			want: netip.MustParseAddr("10.4.71.191"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convert4In6To4(tt.ip, 0)
			assert.Equal(t, tt.want, got)
		})
	}
}
