package hostmonitor

import "math"

func validTemperature(value float64) *float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 || value > 130 {
		return nil
	}
	return &value
}
