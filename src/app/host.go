package app

import (
	"fmt"
	"sync"

	"hatch/hostmonitor"
)

var hostDiskBenchmarkLock sync.Mutex
var hostNetworkBenchmarkLock sync.Mutex
var hostStorageAnalysisLock sync.Mutex

func (a *App) GetHostSnapshot() hostmonitor.Snapshot {
	return hostmonitor.ReadSnapshot()
}

func (a *App) GetHostHardware() (hostmonitor.Hardware, error) {
	return hostmonitor.ReadHardware()
}

func (a *App) GetHostPorts() ([]hostmonitor.Port, error) {
	return hostmonitor.ReadPorts()
}

func (a *App) AnalyzeHostStorage() ([]hostmonitor.StorageEntry, error) {
	if !hostStorageAnalysisLock.TryLock() {
		return nil, fmt.Errorf("磁盘分析正在进行")
	}
	defer hostStorageAnalysisLock.Unlock()
	return hostmonitor.AnalyzeHostStorage()
}

func (a *App) RunHostDiskBenchmark() (hostmonitor.DiskBenchmark, error) {
	if !hostDiskBenchmarkLock.TryLock() {
		return hostmonitor.DiskBenchmark{}, fmt.Errorf("磁盘测速正在进行")
	}
	defer hostDiskBenchmarkLock.Unlock()
	return hostmonitor.BenchmarkDisk()
}

func (a *App) RunHostNetworkBenchmark() (hostmonitor.NetworkQuality, error) {
	if !hostNetworkBenchmarkLock.TryLock() {
		return hostmonitor.NetworkQuality{}, fmt.Errorf("网络测速正在进行")
	}
	defer hostNetworkBenchmarkLock.Unlock()
	return hostmonitor.MeasureNetwork()
}
