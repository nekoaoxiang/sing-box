package provider

import (
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

type Hysteria2Option struct {
	BaseProxy     `yaml:",inline"`
	DialerOptions `yaml:",inline"`
	Ports         string `yaml:"ports,omitempty"`
	HopInterval   string `yaml:"hop-interval,omitempty"`
	Up            int    `yaml:"up,omitempty"`
	Down          int    `yaml:"down,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Obfs          string `yaml:"obfs,omitempty"`
	ObfsPassword  string `yaml:"obfs-password,omitempty"`
	*TLSOption    `yaml:",inline"`

	CustomCA       string `yaml:"ca,omitempty"`
	CustomCAString string `yaml:"ca-str,omitempty"`
}

func newClashHysteria2(tag string, proxy Hysteria2Option) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeHysteria2,
		Tag:  tag,
	}

	options := &option.Hysteria2OutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     proxy.BaseProxy.Server,
			ServerPort: proxy.BaseProxy.Port,
		},
		ServerPorts: convertPortRange(proxy.Ports),
		Password:    proxy.Password,
		Network:     clashNetworks(proxy.BaseProxy.UDP),
		UpMbps:      proxy.Up,
		DownMbps:    proxy.Up,

		DialerOptions: newDialerOptions(proxy.DialerOptions),
	}

	if proxy.Obfs != "" {
		options.Obfs = &option.Hysteria2Obfs{
			Type:     proxy.Obfs,
			Password: proxy.ObfsPassword,
		}
	}

	options.TLS = newTLSOptions(proxy.TLSOption)
	options.TLS.CertificatePath = proxy.CustomCA
	if proxy.CustomCAString != "" {
		options.TLS.Certificate = append(options.TLS.Certificate, proxy.CustomCAString)
	}

	outbound.Options = options
	return outbound, nil
}

func convertPortRange(input string) []string {
	if strings.Contains(input, "-") {
		parts := strings.Split(input, "-")
		if len(parts) == 2 {
			start := strings.TrimSpace(parts[0])
			end := strings.TrimSpace(parts[1])
			return []string{start + ":" + end}
		}
	}
	return nil
}
