package fake

import (
	"net"
	"testing"
)

type Config struct {
	Name string

	IPv4, IPv6, Internal string

	Directories, Files []string
}

func ClientFor(t *testing.T, config *Config) *Client {
	return &Client{
		name:     config.Name,
		ipv4:     net.ParseIP(config.IPv4),
		ipv6:     net.ParseIP(config.IPv6),
		internal: net.ParseIP(config.Internal),
	}
}
