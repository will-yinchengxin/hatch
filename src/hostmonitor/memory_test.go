package hostmonitor

import "testing"

func TestParseMemory(t *testing.T) {
	input := `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free: 10.
Pages active: 20.
Pages inactive: 30.
Pages speculative: 5.
Pages wired down: 40.
Pages occupied by compressor: 8.
`
	used, available, err := parseMemory(input)
	if err != nil || used != 68*16384 || available != 45*16384 {
		t.Fatalf("used=%d available=%d err=%v", used, available, err)
	}
}
