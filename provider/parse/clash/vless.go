package provider

import (
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

type VlessOption struct {
	BaseProxy             `yaml:",inline"`
	DialerOptions         `yaml:",inline"`
	UUID                  string `yaml:"uuid"`
	Flow                  string `yaml:"flow,omitempty"`
	PacketAddr            bool   `yaml:"packet-addr,omitempty"`
	PacketEncoding        string `yaml:"packet-encoding,omitempty"`
	Network               string `yaml:"network,omitempty"`
	*TLSOption            `yaml:",inline"`
	*V2RayTransportOption `yaml:",inline"`
}

func newClashVLESS(tag string, proxy VlessOption) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeVLESS,
		Tag:  tag,
	}
	options := &option.VLESSOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     proxy.BaseProxy.Server,
			ServerPort: proxy.BaseProxy.Port,
		},
	}
	options.UUID = proxy.UUID
	options.Flow = proxy.Flow

	options.TLS = newTLSOptions(proxy.TLSOption)
	options.Transport = newV2RayTransport(proxy.Network, proxy.V2RayTransportOption)
	options.Multiplex = newSMuxOptions(proxy.Smux)
	options.DialerOptions = newDialerOptions(proxy.DialerOptions)

	outbound.Options = options
	return outbound, nil
}
