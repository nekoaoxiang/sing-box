package provider

import (
	"fmt"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func newClashTrojan(proxy map[string]any) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeTrojan,
	}
	options := option.TrojanOutboundOptions{}
	if name, exists := proxy["name"].(string); exists {
		outbound.Tag = name
	}

	if server, exists := proxy["server"].(string); exists {
		options.Server = server
	}
	if port, exists := proxy["port"]; exists {
		options.ServerPort = stringToUint16(fmt.Sprint(port))
	}
	if password, exists := proxy["password"].(string); exists {
		options.Password = password
	}

	options.TLS = newTLSOptions(proxy)
	options.TLS.Enabled = true

	options.Multiplex = newSMuxOptions(proxy)

	options.Transport = newV2RayTransport(proxy)

	options.DialerOptions = newDialerOptions(proxy)

	outbound.Options = options
	return outbound, nil
}
