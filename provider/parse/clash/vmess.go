package provider

import (
	"fmt"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func newClashVMess(proxy map[string]any) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeVMess,
	}

	options := option.VMessOutboundOptions{}
	if name, exists := proxy["name"].(string); exists {
		outbound.Tag = name
	}

	if server, exists := proxy["server"].(string); exists {
		options.Server = server
	}
	if port, exists := proxy["port"]; exists {
		options.ServerPort = stringToUint16(fmt.Sprint(port))
	}
	if uuid, exists := proxy["uuid"].(string); exists {
		options.UUID = uuid
	}
	if aid, exists := proxy["alterId"].(int); exists {
		options.AlterId = aid
	}
	if cipher, exists := proxy["cipher"].(string); exists {
		options.Security = cipher
	}

	options.TLS = newTLSOptions(proxy)
	options.Multiplex = newSMuxOptions(proxy)
	options.Transport = newV2RayTransport(proxy)
	options.DialerOptions = newDialerOptions(proxy)

	outbound.Options = options
	return outbound, nil
}
