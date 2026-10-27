package hostmonitor

import "testing"

func TestCPUUsageFromTicks(t *testing.T) {
	previous := cpuTicks{100, 200, 300, 50}
	current := cpuTicks{120, 220, 350, 60}
	value := cpuUsage(previous, current)
	if value == nil || *value != 50 {
		t.Fatalf("unexpected usage %v", value)
	}
	if cpuUsage(current, current) != nil || cpuUsage(current, previous) != nil {
		t.Fatal("invalid samples must be unavailable")
	}
}
