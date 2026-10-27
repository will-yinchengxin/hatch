package hostmonitor

import "testing"

func TestParsePorts(t *testing.T) {
	input := "p42\ncServer\nf8\nPTCP\nn127.0.0.1:8080\nf9\nPTCP\nn127.0.0.1:8080\nf10\nPUDP\nn*:5353\nf11\nPUDP\nn*:*\n"
	ports := parsePorts(input)
	if len(ports) != 2 || ports[0].PID != 42 || ports[0].Process != "Server" || ports[0].Address != "127.0.0.1:8080" || ports[1].Protocol != "UDP" {
		t.Fatalf("unexpected ports: %#v", ports)
	}
}
