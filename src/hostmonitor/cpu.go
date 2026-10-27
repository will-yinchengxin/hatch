package hostmonitor

type cpuTicks [4]uint64

func cpuUsage(previous, current cpuTicks) *float64 {
	var total, idle uint64
	for index, value := range current {
		if value < previous[index] {
			return nil
		}
		delta := value - previous[index]
		total += delta
		if index == 2 {
			idle = delta
		}
	}
	if total == 0 {
		return nil
	}
	percent := float64(total-idle) * 100 / float64(total)
	return &percent
}
