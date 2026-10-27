//go:build cgo

package hostmonitor

import smc "github.com/caseymrm/go-smc"

func readSensors() (*float64, []int) {
	temperature := validTemperature(smc.ReadTemperature())
	fans := smc.ReadFanSpeeds()
	for _, speed := range fans {
		if speed < 0 || speed > 20000 {
			return temperature, nil
		}
	}
	return temperature, fans
}
