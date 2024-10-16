package provider

import (
	"runtime"

	"github.com/sagernet/sing-box/option"
)

func newDialerOptions(proxy map[string]any) option.DialerOptions {
	options := option.DialerOptions{}

	if detour, exists := proxy["dialer-proxy"].(string); exists {
		options.Detour = detour
	}
	if name, exists := proxy["interface-name"].(string); exists {
		options.BindInterface = name
	}
	if mark, exists := proxy["routing-mark"].(uint32); exists {
		if runtime.GOOS == "android" || runtime.GOOS == "linux" {
			options.RoutingMark = mark
		}
	}
	if tfo, exists := proxy["tcp-fast-open"].(bool); exists {
		options.TCPFastOpen = tfo
	}
	if tfo, exists := proxy["tfo"].(bool); exists {
		options.TCPFastOpen = tfo
	}

	if version, exists := proxy["ip-version"].(string); exists {
		var strategy option.DomainStrategy
		switch version {
		case "dual":
			strategy.UnmarshalJSON([]byte(""))
		case "ipv4":
			strategy.UnmarshalJSON([]byte("ipv4_only"))
		case "ipv6":
			strategy.UnmarshalJSON([]byte("ipv6_only"))
		case "ipv4-prefer":
			strategy.UnmarshalJSON([]byte("prefer_ipv4"))
		case "ipv6-prefer":
			strategy.UnmarshalJSON([]byte("prefer_ipv6"))
		}
		options.DomainStrategy = strategy
	}
	return options
}
