package provider

import (
	"fmt"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func newClashShadowsocks(proxy map[string]any) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeShadowsocks,
	}
	options := &option.ShadowsocksOutboundOptions{}
	UDPOverTCPOptions := &option.UDPOverTCPOptions{}

	if name, exists := proxy["name"].(string); exists {
		outbound.Tag = name
	}
	if server, exists := proxy["server"].(string); exists {
		options.Server = server
	}
	if port, exists := proxy["port"]; exists {
		options.ServerPort = stringToUint16(fmt.Sprint(port))
	}
	if cipher, exists := proxy["cipher"].(string); exists {
		options.Method = cipher
	}
	if password, exists := proxy["password"].(string); exists {
		options.Password = password
	}

	if UDPOverTCP, exists := proxy["udp-over-tcp"].(bool); exists {
		UDPOverTCPOptions.Enabled = UDPOverTCP
	}
	if UDPOverTCPVersion, exists := proxy["udp-over-tcp-version"].(uint8); exists {
		UDPOverTCPOptions.Version = UDPOverTCPVersion
	}
	options.UDPOverTCP = UDPOverTCPOptions

	options.Multiplex = newSMuxOptions(proxy)
	options.DialerOptions = newDialerOptions(proxy)

	outbound.Options = options
	return outbound, nil
}
