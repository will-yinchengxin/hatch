package hostmonitor

import (
	"math"
	"os"
	"testing"
)

func TestValidateTemperature(t *testing.T) {
	for _, value := range []float64{0, -1, 150, math.NaN(), math.Inf(1)} {
		if validTemperature(value) != nil {
			t.Fatalf("invalid temperature %v", value)
		}
	}
	if value := validTemperature(42.5); value == nil || *value != 42.5 {
		t.Fatalf("valid temperature: %v", value)
	}
}

func TestLiveSensors(t *testing.T) {
	if os.Getenv("HATCH_TEST_LIVE_SENSORS") != "1" {
		t.Skip("live sensor test is opt-in")
	}
	temperature, fans := readSensors()
	if temperature == nil {
		t.Fatal("CPU temperature unavailable")
	}
	if len(fans) == 0 {
		t.Fatal("fan speeds unavailable")
	}
	t.Logf("temperature=%.1f C fans=%v rpm", *temperature, fans)
}
