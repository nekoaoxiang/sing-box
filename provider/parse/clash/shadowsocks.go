package provider

import (
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/format"
)

type ShadowSocksOption struct {
	BaseProxy         `yaml:",inline"`
	DialerOptions     `yaml:",inline"`
	Cipher            string         `yaml:"cipher"`
	Password          string         `yaml:"password"`
	UDPOverTCP        bool           `yaml:"udp-over-tcp,omitempty"`
	UDPOverTCPVersion uint8          `yaml:"udp-over-tcp-version,omitempty"`
	Plugin            string         `yaml:"plugin,omitempty"`
	PluginOpts        map[string]any `yaml:"plugin-opts,omitempty"`
	ClientFingerprint string         `yaml:"client-fingerprint,omitempty"`
}

func newClashShadowsocks(tag string, proxy ShadowSocksOption) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeShadowsocks,
		Tag:  tag,
	}

	options := &option.ShadowsocksOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     proxy.BaseProxy.Server,
			ServerPort: proxy.BaseProxy.Port,
		},
		Method:        proxy.Cipher,
		Password:      proxy.Password,
		Plugin:        clashPluginName(proxy.Plugin),
		PluginOptions: clashPluginOptions(proxy.Plugin, proxy.PluginOpts),
		Network:       clashNetworks(proxy.BaseProxy.UDP),
		Multiplex:     newSMuxOptions(proxy.Smux),
		DialerOptions: newDialerOptions(proxy.DialerOptions),
	}

	if UDPOverTCP := proxy.UDPOverTCP; UDPOverTCP {
		options.UDPOverTCP = &option.UDPOverTCPOptions{
			Enabled: proxy.UDPOverTCP,
			Version: proxy.UDPOverTCPVersion,
		}
	}

	outbound.Options = options
	return outbound, nil
}

func clashPluginName(plugin string) string {
	switch plugin {
	case "obfs":
		return "obfs-local"
	}
	return plugin
}

type shadowsocksPluginOptionsBuilder map[string]any

func (o shadowsocksPluginOptionsBuilder) Build() string {
	var opts []string
	for key, value := range o {
		if value == nil {
			continue
		}
		opts = append(opts, format.ToString(key, "=", value))
	}
	return strings.Join(opts, ";")
}

func clashPluginOptions(plugin string, opts map[string]any) string {
	options := make(shadowsocksPluginOptionsBuilder)
	switch plugin {
	case "obfs":
		options["obfs"] = opts["mode"]
		options["obfs-host"] = opts["host"]
	case "v2ray-plugin":
		options["mode"] = opts["mode"]
		options["tls"] = opts["tls"]
		options["host"] = opts["host"]
		options["path"] = opts["path"]
	}
	return options.Build()
}
