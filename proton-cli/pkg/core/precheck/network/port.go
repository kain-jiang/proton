package network

import (
	"fmt"
	"net"
	"strings"
)

var protonPort = []string{
	"6443",
	"2379",
	"10250",
	"10251",
	"10252",
}

func PortAvaiable() [][]string {
	usedPort := []string{}
	portInfo := [][]string{}
	for _, port := range protonPort {
		listener, err := net.Listen("tcp", "0.0.0.0:"+port)
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") {
				usedPort = append(usedPort, port)
			} else {
				portInfo = append(portInfo, []string{"Network Port Available", err.Error(), "\033[31mNO PASS\033[0m", fmt.Sprintf("cannot get port %s status", port)})
			}
			continue
		}
		defer listener.Close()
	}
	if len(usedPort) != 0 {
		portInfo = append(portInfo, []string{"Network Port Available", strings.Join(usedPort, " "), "\033[31mNO PASS\033[0m", "port already in use"})
	} else {
		portInfo = append(portInfo, []string{"Network Port Available", strings.Join(protonPort, " "), "\033[32mPASS\033[0m", ""})
	}
	return portInfo
}
