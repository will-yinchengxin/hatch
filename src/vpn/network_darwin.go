package vpn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

type NetworkAddress struct {
	Interface string `json:"interface"`
	IP        string `json:"ip"`
	Kind      string `json:"kind"`
}

type LocalNetwork struct {
	Addresses        []NetworkAddress `json:"addresses"`
	DefaultInterface string           `json:"defaultInterface"`
	Gateway          string           `json:"gateway"`
	Error            string           `json:"error,omitempty"`
}

type BroadbandAddress struct {
	IP        string `json:"ip"`
	Location  string `json:"location"`
	Interface string `json:"interface"`
	Error     string `json:"error,omitempty"`
}

type ipv4DefaultRoute struct {
	Interface string
	Gateway   netip.Addr
}

const broadbandIPHost = "myip.ipip.net"

var (
	clashTUNRange = netip.MustParsePrefix("198.18.0.0/15")
	sharedIPRange = netip.MustParsePrefix("100.64.0.0/10")
)

// ReadLocalNetwork reads local IPv4 assignments and routing without a network request.
func ReadLocalNetwork() LocalNetwork {
	addresses, err := interfaceIPv4Addresses()
	if err != nil {
		return LocalNetwork{Error: "无法读取本机网卡"}
	}
	route := readIPv4DefaultRoute(context.Background(), "")
	result := LocalNetwork{Addresses: addresses, DefaultInterface: route.Interface}
	if route.Gateway.Is4() {
		result.Gateway = route.Gateway.String()
	}
	return result
}

func interfaceIPv4Addresses() ([]NetworkAddress, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	addresses := make([]NetworkAddress, 0, len(interfaces))
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		assigned, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, raw := range assigned {
			prefix, err := netip.ParsePrefix(raw.String())
			if err != nil || !prefix.Addr().Is4() || !prefix.Addr().IsGlobalUnicast() {
				continue
			}
			ip := prefix.Addr()
			kind := "other"
			switch {
			case isPhysicalInterface(iface.Name):
				kind = "lan"
			case strings.HasPrefix(iface.Name, "bridge"):
				kind = "vm"
			case strings.HasPrefix(iface.Name, "utun"):
				kind = "vpn"
				if iface.Name == "utun1024" || clashTUNRange.Contains(ip) {
					kind = "clash"
				}
			}
			addresses = append(addresses, NetworkAddress{Interface: iface.Name, IP: ip.String(), Kind: kind})
		}
	}
	return addresses, nil
}

func isPhysicalInterface(name string) bool {
	number, ok := strings.CutPrefix(name, "en")
	index, err := strconv.ParseUint(number, 10, 16)
	return ok && err == nil && name == "en"+strconv.FormatUint(index, 10)
}

func readIPv4DefaultRoute(ctx context.Context, scope string) ipv4DefaultRoute {
	ctx, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
	defer cancel()
	args := []string{"-n", "get", "-inet"}
	if scope != "" {
		args = append(args, "-ifscope", scope)
	}
	output, err := exec.CommandContext(ctx, "/sbin/route", append(args, "default")...).Output()
	if err != nil {
		return ipv4DefaultRoute{}
	}
	return parseIPv4DefaultRoute(string(output))
}

func parseIPv4DefaultRoute(output string) ipv4DefaultRoute {
	var route ipv4DefaultRoute
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		switch fields[0] {
		case "interface:":
			route.Interface = fields[1]
		case "gateway:":
			route.Gateway, _ = netip.ParseAddr(fields[1])
		}
	}
	return route
}

func physicalRouteInterface(route ipv4DefaultRoute, addresses []NetworkAddress, scoped func(string) ipv4DefaultRoute) string {
	physical := make(map[string]bool)
	for _, address := range addresses {
		if isPhysicalInterface(address.Interface) {
			physical[address.Interface] = true
		}
	}
	if physical[route.Interface] && route.Gateway.Is4() {
		return route.Interface
	}
	var found string
	for name := range physical {
		candidate := scoped(name)
		if candidate.Interface != name || !candidate.Gateway.Is4() {
			continue
		}
		if found != "" {
			return ""
		}
		found = name
	}
	return found
}

// ReadBroadbandAddress follows the intranet-ip strategy: query a fixed resolver
// and reach IPIP over the selected physical interface without a proxy fallback.
func ReadBroadbandAddress() BroadbandAddress {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	addresses, err := interfaceIPv4Addresses()
	if err != nil {
		return BroadbandAddress{Error: "无法读取物理网卡"}
	}
	name := physicalRouteInterface(readIPv4DefaultRoute(ctx, ""), addresses, func(name string) ipv4DefaultRoute {
		return readIPv4DefaultRoute(ctx, name)
	})
	if name == "" {
		return BroadbandAddress{Error: "未找到明确的 IPv4 物理出口"}
	}
	iface, err := net.InterfaceByName(name)
	if err != nil || iface.Flags&net.FlagUp == 0 {
		return BroadbandAddress{Interface: name, Error: "物理网卡不可用"}
	}
	result, err := probeBroadbandAddress(ctx, iface.Index)
	result.Interface = name
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

func probeBroadbandAddress(ctx context.Context, ifIndex int) (BroadbandAddress, error) {
	// MACOS IP_BOUND_IF makes both DNS and HTTPS use the same IPv4 interface.
	dialer := net.Dialer{Timeout: 3 * time.Second, Control: func(_, _ string, socket syscall.RawConn) error {
		var bindErr error
		if err := socket.Control(func(fd uintptr) {
			bindErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BOUND_IF, ifIndex)
		}); err != nil {
			return err
		}
		return bindErr
	}}
	resolver := net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		if strings.HasPrefix(network, "tcp") {
			network = "tcp4"
		} else {
			network = "udp4"
		}
		return dialer.DialContext(ctx, network, "223.5.5.5:53")
	}}
	dnsCtx, cancelDNS := context.WithTimeout(ctx, 3*time.Second)
	addresses, err := resolver.LookupNetIP(dnsCtx, "ip4", broadbandIPHost+".")
	cancelDNS()
	if err != nil {
		return BroadbandAddress{}, errors.New("物理出口 DNS 查询失败")
	}
	targets := make([]string, 0, len(addresses))
	for _, ip := range addresses {
		if isPublicIPv4(ip) {
			targets = append(targets, net.JoinHostPort(ip.String(), "443"))
		}
	}
	if len(targets) == 0 {
		return BroadbandAddress{}, errors.New("检测服务未返回有效公网地址")
	}
	transport := &http.Transport{
		// The request URL keeps the TLS server name and Host tied to IPIP.
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			for _, target := range targets {
				conn, err := dialer.DialContext(ctx, "tcp4", target)
				if err == nil {
					return conn, nil
				}
			}
			return nil, errors.New("无法建立物理出口连接")
		},
		TLSHandshakeTimeout: 3 * time.Second,
	}
	defer transport.CloseIdleConnections()
	client := http.Client{
		Transport:     transport,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+broadbandIPHost, nil)
	if err != nil {
		return BroadbandAddress{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return BroadbandAddress{}, errors.New("物理网卡无法直连检测服务")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return BroadbandAddress{}, fmt.Errorf("宽带检测服务返回 HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil || len(body) > 4096 {
		return BroadbandAddress{}, errors.New("无法读取宽带检测结果")
	}
	return parseBroadbandIP(string(body))
}

func parseBroadbandIP(body string) (BroadbandAddress, error) {
	text := strings.ReplaceAll(strings.TrimSpace(body), "：", ":")
	label, value, ok := strings.Cut(text, ":")
	fields := strings.Fields(value)
	if !ok || strings.Join(strings.Fields(label), "") != "当前IP" || len(fields) == 0 {
		return BroadbandAddress{}, errors.New("宽带检测结果格式无效")
	}
	ip, err := netip.ParseAddr(fields[0])
	if err != nil || !isPublicIPv4(ip) {
		return BroadbandAddress{}, errors.New("宽带检测结果没有有效的公网 IPv4 地址")
	}
	_, location, _ := strings.Cut(value, "来自于:")
	locationParts := strings.Fields(location)
	var locationDisplay []string
	for _, part := range locationParts {
		if len(locationDisplay) == 0 || locationDisplay[len(locationDisplay)-1] != part {
			locationDisplay = append(locationDisplay, part)
		}
	}
	return BroadbandAddress{IP: ip.String(), Location: strings.Join(locationDisplay, " ")}, nil
}

func isPublicIPv4(ip netip.Addr) bool {
	return ip.Is4() && ip.IsGlobalUnicast() && !ip.IsPrivate() &&
		!clashTUNRange.Contains(ip) && !sharedIPRange.Contains(ip)
}
