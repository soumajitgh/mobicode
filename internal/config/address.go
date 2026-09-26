package config

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

func normalizeBaseURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", fmt.Errorf("%s must be an http(s) origin without a path, query, or fragment", EnvServerBaseURL)
	}
	if strings.EqualFold(u.Hostname(), "localhost") {
		return "", fmt.Errorf("%s must identify a reachable server", EnvServerBaseURL)
	}
	if ip := net.ParseIP(u.Hostname()); ip != nil && (ip.IsLoopback() || ip.IsUnspecified()) {
		return "", fmt.Errorf("%s must identify a reachable server", EnvServerBaseURL)
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return "", fmt.Errorf("%s has an invalid port", EnvServerBaseURL)
		}
	}
	u.Path = ""
	u.RawPath = ""
	return strings.TrimSuffix(u.String(), "/"), nil
}

func discoverBaseURL(port int) string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	sort.Slice(interfaces, func(i, j int) bool {
		a, b := interfacePriority(interfaces[i].Name), interfacePriority(interfaces[j].Name)
		if a != b {
			return a < b
		}
		return interfaces[i].Name < interfaces[j].Name
	})
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		var candidates []string
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err == nil && ip.To4() != nil && ip.IsPrivate() && !ip.IsLoopback() {
				candidates = append(candidates, ip.String())
			}
		}
		if len(candidates) > 0 {
			sort.Strings(candidates)
			return fmt.Sprintf("http://%s:%d", candidates[0], port)
		}
	}
	return ""
}

func interfacePriority(name string) int {
	for _, prefix := range []string{"en", "eth", "wlan", "wl"} {
		if strings.HasPrefix(name, prefix) {
			return 0
		}
	}
	for _, prefix := range []string{"docker", "br-", "veth", "virbr", "utun", "tun", "tap"} {
		if strings.HasPrefix(name, prefix) {
			return 2
		}
	}
	return 1
}
