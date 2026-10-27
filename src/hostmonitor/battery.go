package hostmonitor

import (
	"regexp"
	"strconv"
	"strings"
)

var batteryPattern = regexp.MustCompile(`(\d+)%;\s*([^;\n]+)`)

func parseBattery(output string) (int, string, bool) {
	match := batteryPattern.FindStringSubmatch(output)
	if len(match) != 3 {
		return 0, "", false
	}
	percent, err := strconv.Atoi(match[1])
	if err != nil || percent > 100 {
		return 0, "", false
	}
	return percent, strings.TrimSpace(match[2]), true
}
