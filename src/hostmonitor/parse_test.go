package hostmonitor

import "testing"

func TestParseNetworkCounters(t *testing.T) {
	input := `Name Mtu Network Address Ipkts Ierrs Ibytes Opkts Oerrs Obytes Coll
lo0 16384 <Link#1> 10 0 100 10 0 100 0
en0 1500 <Link#2> aa:bb:cc:dd:ee:ff 42 0 12345 32 0 6789 0
en0 1500 192.168 192.168.1.2 42 - 12345 32 - 6789 -
utun0 1380 <Link#3> 1 0 99 1 0 88 0
`
	counters, err := parseNetworkCounters(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(counters) != 2 || counters[0].Name != "en0" || counters[0].Received != 12345 || counters[0].Sent != 6789 || counters[1].Name != "utun0" {
		t.Fatalf("unexpected counters: %#v", counters)
	}
}

func TestRatesRejectCounterResetAndZeroInterval(t *testing.T) {
	if rate(10, 5, 1) != 0 || rate(5, 10, 0) != 0 {
		t.Fatal("invalid samples must not produce a rate")
	}
	if got := rate(100, 500, 2); got != 200 {
		t.Fatalf("got %v", got)
	}
}
