package network

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func findDefaultGatewayForInterface(iface *net.Interface) (net.IP, error) {
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_V4, &netlink.Route{LinkIndex: iface.Index}, netlink.RT_FILTER_OIF)
	if err != nil {
		return nil, err
	}

	for _, route := range routes {
		if route.Dst == nil || route.Dst.String() == "0.0.0.0/0" { // Default route
			return route.Gw, nil
		}
	}

	return nil, nil
}

func GatewayTest() [][]string {
	gwInfo := [][]string{}
	interfaces, err := net.Interfaces()
	if err != nil {
		gwInfo = append(gwInfo, []string{"Network Default GW", err.Error(), "\033[31mNO PASS\033[0m", "check network interfaces"})
		return gwInfo
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Printf("ERROR: failed to get addresses for interface %s: %v\n", iface.Name, err)
			continue
		}

		for _, addr := range addrs {
			// Check for IPNet type address which includes the subnet mask
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil { // IPv4 only
					defaultGateway, _ := findDefaultGatewayForInterface(&iface)
					if defaultGateway != nil {
						gwInfo = append(gwInfo, []string{"Network Default GW", iface.Name + ":" + defaultGateway.String(), "\033[32mPASS\033[0m", ""})
						break
					}
				}
			}
		}
	}

	return gwInfo
}
