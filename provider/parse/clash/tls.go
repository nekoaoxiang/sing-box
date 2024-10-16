package provider

import (
	"fmt"

	"github.com/sagernet/sing-box/option"
)

func newTLSOptions(proxy map[string]any) *option.OutboundTLSOptions {
	options := &option.OutboundTLSOptions{
		ECH:     &option.OutboundECHOptions{},
		UTLS:    &option.OutboundUTLSOptions{},
		Reality: &option.OutboundRealityOptions{},
	}
	if tls, exists := proxy["tls"].(bool); exists {
		options.Enabled = tls
	}
	if disableSNI, exists := proxy["disable-sni"].(bool); exists {
		options.DisableSNI = disableSNI
	}
	if sni, exists := proxy["sni"].(string); exists {
		options.ServerName = sni
	}
	if peer, exists := proxy["peer"].(string); exists {
		options.ServerName = peer
	}
	if servername, exists := proxy["servername"].(string); exists {
		options.ServerName = servername
	}
	if insecure, exists := proxy["skip-cert-verify"].(bool); exists {
		options.Enabled = true
		options.Insecure = insecure
	}
	if alpn, exists := proxy["alpn"].([]any); exists {
		alpnArr := []string{}
		for _, item := range alpn {
			alpnArr = append(alpnArr, fmt.Sprint(item))
		}
		options.ALPN = alpnArr
	}
	if fingerprint, exists := proxy["client-fingerprint"].(string); exists {
		options.Enabled = true
		options.UTLS.Enabled = true
		options.UTLS.Fingerprint = fingerprint
	}
	if reality, exists := proxy["reality-opts"].(map[string]any); exists {
		options.Enabled = true
		options.Reality.Enabled = true
		if pbk, exists := reality["public-key"].(string); exists {
			options.Reality.PublicKey = pbk
		}
		if sid, exists := reality["short-id"].(string); exists {
			options.Reality.ShortID = sid
		}
	}
	return options
}
