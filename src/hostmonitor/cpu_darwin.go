//go:build cgo

package hostmonitor

/*
#include <mach/mach.h>
static kern_return_t read_cpu_ticks(host_cpu_load_info_data_t *load) {
    mach_msg_type_number_t count = HOST_CPU_LOAD_INFO_COUNT;
    host_t host = mach_host_self();
    kern_return_t result = host_statistics(host, HOST_CPU_LOAD_INFO, (host_info_t)load, &count);
    mach_port_deallocate(mach_task_self(), host);
    return result;
}
*/
import "C"

import "sync"

var cpuLock sync.Mutex
var previousCPUTicks *cpuTicks

func readCPU() *float64 {
	cpuLock.Lock()
	defer cpuLock.Unlock()
	var load C.host_cpu_load_info_data_t
	if C.read_cpu_ticks(&load) != C.KERN_SUCCESS {
		return nil
	}
	current := cpuTicks{uint64(load.cpu_ticks[0]), uint64(load.cpu_ticks[1]), uint64(load.cpu_ticks[2]), uint64(load.cpu_ticks[3])}
	previous := previousCPUTicks
	previousCPUTicks = &current
	if previous == nil {
		return nil
	}
	return cpuUsage(*previous, current)
}
