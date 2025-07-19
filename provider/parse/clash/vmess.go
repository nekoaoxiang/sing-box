package provider

import (
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

type VMessOption struct {
	BaseProxy             `yaml:",inline"`
	DialerOptions         `yaml:",inline"`
	UUID                  string `yaml:"uuid"`
	AlterID               int    `yaml:"alterId"`
	Cipher                string `yaml:"cipher"`
	Network               string `yaml:"network,omitempty"`
	*TLSOption            `yaml:",inline"`
	PacketAddr            bool   `yaml:"packet-addr,omitempty"`
	PacketEncoding        string `yaml:"packet-encoding,omitempty"`
	GlobalPadding         bool   `yaml:"global-padding,omitempty"`
	AuthenticatedLength   bool   `yaml:"authenticated-length,omitempty"`
	*V2RayTransportOption `yaml:",inline"`
}

func newClashVMess(tag string, proxy VMessOption) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeVMess,
		Tag:  tag,
	}
	options := &option.VMessOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     proxy.BaseProxy.Server,
			ServerPort: proxy.BaseProxy.Port,
		},
		UUID:                proxy.UUID,
		Security:            proxy.Cipher,
		AlterId:             proxy.AlterID,
		GlobalPadding:       proxy.GlobalPadding,
		AuthenticatedLength: proxy.AuthenticatedLength,
		Network:             clashNetworks(proxy.BaseProxy.UDP),
		PacketEncoding:      proxy.PacketEncoding,
	}

	options.TLS = newTLSOptions(proxy.TLSOption)
	options.Multiplex = newSMuxOptions(proxy.Smux)
	options.Transport = newV2RayTransport(proxy.Network, proxy.V2RayTransportOption)
	options.DialerOptions = newDialerOptions(proxy.DialerOptions)

	outbound.Options = options
	return outbound, nil
}
