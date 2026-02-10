package fake

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientFor(t *testing.T) {
	config := &Config{
		Name:     "example",
		IPv4:     "10.4.14.71",
		IPv6:     "fe80::e487:43a7:afea:e959",
		Internal: "10.10.14.71",
	}

	client := ClientFor(t, config)

	assert.Equal(t, net.IPv4(10, 4, 14, 71), client.ipv4)
	assert.Equal(t, net.ParseIP("fe80::e487:43a7:afea:e959"), client.ipv6)
	assert.Equal(t, net.IPv4(10, 10, 14, 71), client.internal)
}
