package hostmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

type Snapshot struct {
	Timestamp       int64            `json:"timestamp"`
	CPUPercent      *float64         `json:"cpuPercent"`
	CPUTemperature  *float64         `json:"cpuTemperature"`
	FanSpeeds       []int            `json:"fanSpeeds"`
	MemoryUsed      uint64           `json:"memoryUsed"`
	MemoryAvailable uint64           `json:"memoryAvailable"`
	MemoryTotal     uint64           `json:"memoryTotal"`
	DiskTotal       uint64           `json:"diskTotal"`
	DiskAvailable   uint64           `json:"diskAvailable"`
	Networks        []NetworkCounter `json:"networks"`
	BatteryPercent  *int             `json:"batteryPercent"`
	BatteryStatus   string           `json:"batteryStatus"`
	Errors          []string         `json:"errors"`
}

func command(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).Output()
}

func ReadSnapshot() Snapshot {
	snapshot := Snapshot{Timestamp: time.Now().UnixMilli(), Errors: []string{}, Networks: []NetworkCounter{}}
	snapshot.CPUTemperature, snapshot.FanSpeeds = readSensors()
	if data, err := command(3*time.Second, "/usr/bin/pmset", "-g", "batt"); err == nil {
		if percent, status, ok := parseBattery(string(data)); ok {
			snapshot.BatteryPercent, snapshot.BatteryStatus = &percent, status
		}
	}
	if data, err := command(5*time.Second, "/usr/bin/vm_stat"); err == nil {
		if used, available, parseErr := parseMemory(string(data)); parseErr == nil {
			snapshot.MemoryUsed, snapshot.MemoryAvailable = used, available
		} else {
			snapshot.Errors = append(snapshot.Errors, "内存: "+parseErr.Error())
		}
	} else {
		snapshot.Errors = append(snapshot.Errors, "内存采集失败")
	}
	if data, err := command(3*time.Second, "/usr/sbin/sysctl", "-n", "hw.memsize"); err == nil {
		snapshot.MemoryTotal, _ = strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	}
	var disk unix.Statfs_t
	if err := unix.Statfs("/System/Volumes/Data", &disk); err == nil {
		snapshot.DiskTotal = disk.Blocks * uint64(disk.Bsize)
		snapshot.DiskAvailable = disk.Bavail * uint64(disk.Bsize)
	} else {
		snapshot.Errors = append(snapshot.Errors, "磁盘容量采集失败")
	}
	if data, err := command(4*time.Second, "/usr/sbin/netstat", "-ibn"); err == nil {
		if counters, parseErr := parseNetworkCounters(string(data)); parseErr == nil {
			snapshot.Networks = counters
		} else {
			snapshot.Errors = append(snapshot.Errors, "网络: "+parseErr.Error())
		}
	} else {
		snapshot.Errors = append(snapshot.Errors, "网络采集失败")
	}
	snapshot.CPUPercent = readCPU()
	return snapshot
}

type Hardware struct {
	Model          string `json:"model"`
	Chip           string `json:"chip"`
	Cores          string `json:"cores"`
	Memory         string `json:"memory"`
	GPU            string `json:"gpu"`
	Display        string `json:"display"`
	BatteryCycles  string `json:"batteryCycles"`
	BatteryHealth  string `json:"batteryHealth"`
	BatteryMaximum string `json:"batteryMaximum"`
	OS             string `json:"os"`
}

func ReadHardware() (Hardware, error) {
	data, err := command(25*time.Second, "/usr/sbin/system_profiler", "SPHardwareDataType", "SPDisplaysDataType", "SPPowerDataType", "-json")
	if err != nil {
		return Hardware{}, fmt.Errorf("硬件信息不可用: %w", err)
	}
	var root map[string][]map[string]any
	if err = json.Unmarshal(data, &root); err != nil {
		return Hardware{}, err
	}
	value := func(object map[string]any, key string) string {
		if v, ok := object[key]; ok {
			return fmt.Sprint(v)
		}
		return ""
	}
	result := Hardware{}
	if list := root["SPHardwareDataType"]; len(list) > 0 {
		result.Model = value(list[0], "machine_name")
		result.Chip = value(list[0], "chip_type")
		result.Cores = value(list[0], "number_processors")
		result.Memory = value(list[0], "physical_memory")
	}
	if list := root["SPDisplaysDataType"]; len(list) > 0 {
		result.GPU = value(list[0], "sppci_model")
		if displays, ok := list[0]["spdisplays_ndrvs"].([]any); ok && len(displays) > 0 {
			if display, ok := displays[0].(map[string]any); ok {
				result.Display = value(display, "_spdisplays_pixels")
			}
		}
	}
	if list := root["SPPowerDataType"]; len(list) > 0 {
		if health, ok := list[0]["sppower_battery_health_info"].(map[string]any); ok {
			result.BatteryCycles = value(health, "sppower_battery_cycle_count")
			result.BatteryHealth = value(health, "sppower_battery_health")
			result.BatteryMaximum = value(health, "sppower_battery_health_maximum_capacity")
		}
	}
	if version, versionErr := command(3*time.Second, "/usr/bin/sw_vers", "-productVersion"); versionErr == nil {
		result.OS = strings.TrimSpace(string(version))
	}
	return result, nil
}
