package provider

import (
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

type TLSOption struct {
	TLS               bool            `yaml:"tls,omitempty"`
	SNI               string          `yaml:"sni,omitempty"`
	ServerName        string          `yaml:"servername,omitempty"`
	Fingerprint       string          `yaml:"fingerprint,omitempty"`
	ALPN              []string        `yaml:"alpn,omitempty"`
	SkipCertVerify    bool            `yaml:"skip-cert-verify,omitempty"`
	ClientFingerprint string          `yaml:"client-fingerprint,omitempty"`
	RealityOpts       *RealityOptions `yaml:"reality-opts,omitempty"`
	ECHOpts           *ECHOptions     `yaml:"ech-opts,omitempty"`
}

type RealityOptions struct {
	PublicKey string `yaml:"public-key,omitempty"`
	ShortID   string `yaml:"short-id,omitempty"`
}

type ECHOptions struct {
	Enabled bool   `yaml:"enable,omitempty"`
	Config  string `yaml:"config,omitempty"`
}

func newTLSOptions(proxy *TLSOption) *option.OutboundTLSOptions {
	if proxy == nil {
		return nil
	}

	var options option.OutboundTLSOptions

	options.ServerName = proxy.ServerName
	if options.ServerName == "" {
		options.ServerName = proxy.SNI
	}

	options.Enabled = proxy.TLS || proxy.SkipCertVerify || options.ServerName != "" || proxy.RealityOpts != nil

	options.Insecure = proxy.SkipCertVerify

	if len(proxy.ALPN) > 0 {
		options.ALPN = append(options.ALPN, proxy.ALPN...)
	}

	if proxy.ECHOpts != nil {
		options.ECH = &option.OutboundECHOptions{
			Enabled: proxy.ECHOpts.Enabled,
			Config:  badoption.Listable[string](append([]string{}, proxy.ECHOpts.Config)),
		}
	}

	if proxy.Fingerprint != "" {
		options.UTLS = &option.OutboundUTLSOptions{
			Enabled:     true,
			Fingerprint: proxy.Fingerprint,
		}
	}

	if proxy.RealityOpts != nil {
		options.Reality = &option.OutboundRealityOptions{
			Enabled:   true,
			PublicKey: proxy.RealityOpts.PublicKey,
			ShortID:   proxy.RealityOpts.ShortID,
		}
	}
	return &options
}
