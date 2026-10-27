//go:build !cgo

package hostmonitor

func readSensors() (*float64, []int) { return nil, nil }
