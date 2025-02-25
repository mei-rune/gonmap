package gonmap

import (
	"fmt"
	"kscan/lib/ping"
	"time"
)

func HostDiscovery(ip string) bool {
	if HostDiscoveryForIcmp(ip) {
		return true
	}
	if HostDiscoveryForTcp(ip) {
		return true
	}
	return false
}

func HostDiscoveryForIcmp(ip string) bool {
	if ping.Check(ip) {
		return true
	}
	return false
}

func HostDiscoveryForTcp(ip string) bool {
	tcpArr := []int{22, 23, 80, 139, 443, 445, 3389}
	for _, port := range tcpArr {
		netloc := fmt.Sprintf("%s:%d", ip, port)
		if PortScan("tcp", netloc, time.Second*2) {
			return true
		}
	}
	return false
}
