package hostmonitor

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Port struct {
	PID      int    `json:"pid"`
	Process  string `json:"process"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
}

func parsePorts(output string) []Port {
	ports := make([]Port, 0)
	seen := make(map[Port]bool)
	var pid int
	var process, protocol string
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(line[1:])
			process, protocol = "", ""
		case 'c':
			process = line[1:]
		case 'f':
			protocol = ""
		case 'P':
			protocol = line[1:]
		case 'n':
			address := line[1:]
			if pid <= 0 || process == "" || (protocol != "TCP" && protocol != "UDP") || strings.HasSuffix(address, ":*") {
				continue
			}
			entry := Port{PID: pid, Process: process, Protocol: protocol, Address: address}
			if !seen[entry] {
				seen[entry] = true
				ports = append(ports, entry)
			}
		}
	}
	return ports
}

func ReadPorts() ([]Port, error) {
	data, err := command(20*time.Second, "/usr/sbin/lsof", "-nP", "-i", "-sTCP:LISTEN", "-FpcPn")
	if err != nil {
		return nil, fmt.Errorf("端口信息不可用或采集超时: %w", err)
	}
	return parsePorts(string(data)), nil
}
