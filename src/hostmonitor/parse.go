package hostmonitor

import (
	"fmt"
	"strconv"
	"strings"
)

type NetworkCounter struct {
	Name     string `json:"name"`
	Received uint64 `json:"received"`
	Sent     uint64 `json:"sent"`
}

func parseNetworkCounters(output string) ([]NetworkCounter, error) {
	var counters []NetworkCounter
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 10 || !strings.HasPrefix(fields[2], "<Link#") || strings.HasPrefix(fields[0], "lo") || strings.HasSuffix(fields[0], "*") {
			continue
		}
		received, err := strconv.ParseUint(fields[len(fields)-5], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("network receive counter: %w", err)
		}
		sent, err := strconv.ParseUint(fields[len(fields)-2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("network send counter: %w", err)
		}
		counters = append(counters, NetworkCounter{Name: fields[0], Received: received, Sent: sent})
	}
	if len(counters) == 0 {
		return nil, fmt.Errorf("no network interfaces found")
	}
	return counters, nil
}

func rate(previous, current uint64, seconds float64) float64 {
	if current < previous || seconds <= 0 {
		return 0
	}
	return float64(current-previous) / seconds
}
