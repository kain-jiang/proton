package v1alpha1

import (
	"bufio"
	"io"
	"net"
	"regexp"
	"sort"
	"strconv"
)

type NetworkInterface struct {
	Index int

	Name string

	Addresses []net.IPNet
}

var (
	linkRegex  = regexp.MustCompile(`(^\d+): ([a-z][0-9a-zA-Z]+\.{0,1}\d{0,})@{0,}[a-z]{0,}[0-9a-zA-Z@]{0,}:`)
	inetRegex  = regexp.MustCompile(`^    inet (.*/\d+) `)
	inet6Regex = regexp.MustCompile(`^    inet6 (.*/\d+) `)
)

func parseOutputOfIPAddress(r io.Reader) ([]NetworkInterface, error) {
	scanner := bufio.NewScanner(r)

	var interfaces []NetworkInterface

	var ifi *NetworkInterface
	for scanner.Scan() {
		line := scanner.Text()

		if matches := linkRegex.FindStringSubmatch(line); matches != nil {
			if ifi != nil {
				interfaces = append(interfaces, *ifi)
			}
			ifi = new(NetworkInterface)

			i, err := strconv.Atoi(matches[1])
			if err != nil {
				return nil, err
			}
			ifi.Index = i
			ifi.Name = matches[2]
			continue
		}

		if matches := inetRegex.FindStringSubmatch(line); matches != nil {
			ip, ipNet, err := net.ParseCIDR(matches[1])
			if err != nil {
				return nil, err
			}
			ifi.Addresses = append(ifi.Addresses, net.IPNet{IP: ip, Mask: ipNet.Mask})
			continue
		}

		if matches := inet6Regex.FindStringSubmatch(line); matches != nil {
			ip, ipNet, err := net.ParseCIDR(matches[1])
			if err != nil {
				return nil, err
			}
			ifi.Addresses = append(ifi.Addresses, net.IPNet{IP: ip, Mask: ipNet.Mask})
			continue
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	if ifi != nil {
		interfaces = append(interfaces, *ifi)
	}

	// sort network interfaces by index
	sort.Slice(interfaces, func(i, j int) bool {
		return interfaces[i].Index < interfaces[j].Index
	})

	return interfaces, nil
}
