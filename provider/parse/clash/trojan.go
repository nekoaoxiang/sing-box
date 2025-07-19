package provider

import (
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

type TrojanOption struct {
	BaseProxy             `yaml:",inline"`
	DialerOptions         `yaml:",inline"`
	Password              string `yaml:"password"`
	*TLSOption            `yaml:",inline"`
	Network               string `yaml:"network,omitempty"`
	*V2RayTransportOption `yaml:",inline"`
}

func newClashTrojan(tag string, proxy TrojanOption) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeTrojan,
		Tag:  tag,
	}
	options := &option.TrojanOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     proxy.BaseProxy.Server,
			ServerPort: proxy.BaseProxy.Port,
		},
		Password: proxy.Password,
	}

	options.TLS = newTLSOptions(proxy.TLSOption)
	options.Multiplex = newSMuxOptions(proxy.Smux)
	options.Transport = newV2RayTransport(proxy.Network, proxy.V2RayTransportOption)
	options.DialerOptions = newDialerOptions(proxy.DialerOptions)

	outbound.Options = options
	return outbound, nil
}
