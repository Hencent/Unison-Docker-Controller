package pkg

import (
	"github.com/docker/go-connections/nat"
	"strconv"
)

func generateExportsForContainer(tcpList []int, udpList []int) (nat.PortSet, error) {
	exports := make(nat.PortSet)
	for tcp := range tcpList {
		port, err := nat.NewPort("tcp", strconv.Itoa(tcp))
		if err != nil {
			return nil, err
		}
		exports[port] = struct{}{}
	}
	for udp := range udpList {
		port, err := nat.NewPort("udp", strconv.Itoa(udp))
		if err != nil {
			return nil, err
		}
		exports[port] = struct{}{}
	}

	return exports, nil
}
