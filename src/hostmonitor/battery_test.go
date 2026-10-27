package hostmonitor

import "testing"

func TestParseBattery(t *testing.T) {
	percent, status, ok := parseBattery("Now drawing from 'Battery Power'\n -InternalBattery-0 (id=1)\t59%; discharging; 2:42 remaining present: true\n")
	if !ok || percent != 59 || status != "discharging" {
		t.Fatalf("%v %q %v", percent, status, ok)
	}
	if _, _, ok = parseBattery("No batteries"); ok {
		t.Fatal("missing battery should be unavailable")
	}
}
