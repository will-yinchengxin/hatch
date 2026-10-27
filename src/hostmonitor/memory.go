package hostmonitor

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var pageSizePattern = regexp.MustCompile(`page size of (\d+) bytes`)

func parseMemory(output string) (uint64, uint64, error) {
	match := pageSizePattern.FindStringSubmatch(output)
	if len(match) != 2 {
		return 0, 0, fmt.Errorf("memory page size unavailable")
	}
	pageSize, err := strconv.ParseUint(match[1], 10, 64)
	if err != nil || pageSize == 0 {
		return 0, 0, fmt.Errorf("invalid memory page size")
	}
	values := make(map[string]uint64)
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		value, parseErr := strconv.ParseUint(strings.Trim(strings.TrimSpace(parts[1]), "."), 10, 64)
		if parseErr == nil {
			values[strings.TrimSpace(parts[0])] = value
		}
	}
	for _, key := range []string{"Pages free", "Pages active", "Pages inactive", "Pages speculative", "Pages wired down", "Pages occupied by compressor"} {
		if _, ok := values[key]; !ok {
			return 0, 0, fmt.Errorf("memory field %q unavailable", key)
		}
	}
	used := values["Pages active"] + values["Pages wired down"] + values["Pages occupied by compressor"]
	available := values["Pages free"] + values["Pages inactive"] + values["Pages speculative"]
	return used * pageSize, available * pageSize, nil
}
