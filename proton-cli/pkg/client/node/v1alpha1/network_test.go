package v1alpha1

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseOutputOfIPAddress(t *testing.T) {
	f, err := os.Open("testdata/output_ip_address.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var want = []NetworkInterface{
		{
			Index: 1,
			Name:  "lo",
			Addresses: []net.IPNet{
				{
					IP:   net.ParseIP("127.0.0.1"),
					Mask: net.CIDRMask(8, 32),
				},
				{
					IP:   net.ParseIP("::1"),
					Mask: net.CIDRMask(128, 128),
				},
			},
		},
		{
			Index: 2,
			Name:  "ens160",
			Addresses: []net.IPNet{
				{
					IP:   net.ParseIP("10.4.14.71"),
					Mask: net.CIDRMask(24, 32),
				},
				{
					IP:   net.ParseIP("fe80::e487:43a7:afea:e959"),
					Mask: net.CIDRMask(64, 128),
				},
			},
		},
		{
			Index: 3,
			Name:  "ens192",
			Addresses: []net.IPNet{
				{
					IP:   net.ParseIP("10.10.14.71"),
					Mask: net.CIDRMask(24, 32),
				},
				{
					IP:   net.ParseIP("fe80::f882:bb69:4774:29a9"),
					Mask: net.CIDRMask(64, 128),
				},
			},
		},
		{
			Index: 4,
			Name:  "docker0",
			Addresses: []net.IPNet{
				{
					IP:   net.ParseIP("172.33.0.1"),
					Mask: net.CIDRMask(16, 32),
				},
			},
		},
		{
			Index: 5,
			Name:  "tunl0",
			Addresses: []net.IPNet{
				{
					IP:   net.ParseIP("192.169.223.0"),
					Mask: net.CIDRMask(32, 32),
				},
			},
		},
	}

	got, err := parseOutputOfIPAddress(f)
	if assert.NoError(t, err) {
		assert.Equal(t, got, want)
	}
}
