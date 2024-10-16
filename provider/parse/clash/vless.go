package provider

import (
	"fmt"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func newClashVLESS(proxy map[string]any) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeVLESS,
	}
	options := option.VLESSOutboundOptions{}
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
	if flow, exists := proxy["flow"].(string); exists && flow == "xtls-rprx-vision" {
		options.Flow = "xtls-rprx-vision"
	}

	options.TLS = newTLSOptions(proxy)

	options.Transport = newV2RayTransport(proxy)

	options.Multiplex = newSMuxOptions(proxy)
	options.DialerOptions = newDialerOptions(proxy)
	outbound.Options = options
	return outbound, nil
}
