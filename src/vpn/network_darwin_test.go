package vpn

import (
	"net/netip"
	"testing"
)

func TestBroadbandAddressResponse(t *testing.T) {
	address, err := parseBroadbandIP("当前 IP：180.154.58.92  来自于：中国 上海 上海 电信\n")
	if err != nil || address.IP != "180.154.58.92" || address.Location != "中国 上海 电信" {
		t.Fatalf("address=%+v err=%v", address, err)
	}
	for _, input := range []string{
		"error: 180.154.58.92",
		"当前 IP：198.18.0.1 来自于：Clash",
		"当前 IP：192.168.50.10 来自于：中国",
	} {
		if _, err := parseBroadbandIP(input); err == nil {
			t.Fatalf("unexpected public address in %q", input)
		}
	}
}

func TestPhysicalRouteSelection(t *testing.T) {
	physical := []NetworkAddress{
		{Interface: "en0", IP: "192.168.50.116"},
		{Interface: "en5", IP: "192.168.60.116"},
		{Interface: "utun1024", IP: "198.18.0.1"},
	}
	tunnel := parseIPv4DefaultRoute("   gateway: 198.18.0.1\n  interface: utun1024\n")
	if tunnel.Interface != "utun1024" || tunnel.Gateway != netip.MustParseAddr("198.18.0.1") {
		t.Fatalf("unexpected route %+v", tunnel)
	}
	defaultRoute := ipv4DefaultRoute{Interface: "en0", Gateway: netip.MustParseAddr("192.168.50.1")}
	routes := map[string]ipv4DefaultRoute{"en0": defaultRoute}
	scoped := func(name string) ipv4DefaultRoute { return routes[name] }
	if iface := physicalRouteInterface(tunnel, physical, scoped); iface != "en0" {
		t.Fatalf("wanted unique physical route, got %q", iface)
	}
	routes["en5"] = ipv4DefaultRoute{Interface: "en5", Gateway: netip.MustParseAddr("192.168.60.1")}
	if iface := physicalRouteInterface(tunnel, physical, scoped); iface != "" {
		t.Fatalf("ambiguous fallback chose %q", iface)
	}
	if iface := physicalRouteInterface(defaultRoute, physical, scoped); iface != "en0" {
		t.Fatalf("primary physical route chose %q", iface)
	}
}
