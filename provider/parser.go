package provider

import (
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

func (p *myProviderAdapter) newParser(content string) ([]option.Outbound, error) {
	var outbounds []option.Outbound
	var err error
	switch true {
	case strings.Contains(content, "\"outbounds\""):
		var options option.OutboundProviderOptions
		err = options.UnmarshalJSON([]byte(content))
		if err != nil {
			return nil, E.Cause(err, "decode config at ")
		}
		outbounds = options.Outbounds
	case strings.Contains(content, "proxies"):
		outbounds, err = newClashParser(content)
		if err != nil {
			return nil, err
		}
	default:
		outbounds, err = newNativeURIParser(content)
		if err != nil {
			return nil, err
		}
	}
	return p.overrideOutbounds(outbounds), nil
}

func (p *myProviderAdapter) overrideOutbounds(outbounds []option.Outbound) []option.Outbound {
	if p.overrideDialer == nil {
		return outbounds
	}
	var parsedOutbounds []option.Outbound
	for _, outbound := range outbounds {
		switch outbound.Type {
		case C.TypeHTTP:
			dialer := outbound.HTTPOptions.DialerOptions
			outbound.HTTPOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeSOCKS:
			dialer := outbound.SocksOptions.DialerOptions
			outbound.SocksOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeTUIC:
			dialer := outbound.TUICOptions.DialerOptions
			outbound.TUICOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeVMess:
			dialer := outbound.VMessOptions.DialerOptions
			outbound.VMessOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeVLESS:
			dialer := outbound.VLESSOptions.DialerOptions
			outbound.VLESSOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeTrojan:
			dialer := outbound.TrojanOptions.DialerOptions
			outbound.TrojanOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeHysteria:
			dialer := outbound.HysteriaOptions.DialerOptions
			outbound.HysteriaOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeShadowTLS:
			dialer := outbound.ShadowTLSOptions.DialerOptions
			outbound.ShadowTLSOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeHysteria2:
			dialer := outbound.Hysteria2Options.DialerOptions
			outbound.Hysteria2Options.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeWireGuard:
			dialer := outbound.WireGuardOptions.DialerOptions
			outbound.WireGuardOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeShadowsocks:
			dialer := outbound.ShadowsocksOptions.DialerOptions
			outbound.ShadowsocksOptions.DialerOptions = p.overrideDialerOption(dialer)
		case C.TypeShadowsocksR:
			dialer := outbound.ShadowsocksROptions.DialerOptions
			outbound.ShadowsocksROptions.DialerOptions = p.overrideDialerOption(dialer)
		}
		parsedOutbounds = append(parsedOutbounds, outbound)
	}
	return parsedOutbounds
}

func (p *myProviderAdapter) overrideDialerOption(options option.DialerOptions) option.DialerOptions {
	if p.overrideDialer.Detour != nil && options.Detour != "" {
		options.Detour = *p.overrideDialer.Detour
	}
	if p.overrideDialer.BindInterface != nil {
		options.BindInterface = *p.overrideDialer.BindInterface
	}
	if p.overrideDialer.Inet4BindAddress != nil {
		options.Inet4BindAddress = p.overrideDialer.Inet4BindAddress
	}
	if p.overrideDialer.Inet6BindAddress != nil {
		options.Inet6BindAddress = p.overrideDialer.Inet6BindAddress
	}
	if p.overrideDialer.ProtectPath != nil {
		options.ProtectPath = *p.overrideDialer.ProtectPath
	}
	if p.overrideDialer.RoutingMark != nil {
		options.RoutingMark = *p.overrideDialer.RoutingMark
	}
	if p.overrideDialer.ReuseAddr != nil {
		options.ReuseAddr = *p.overrideDialer.ReuseAddr
	}
	if p.overrideDialer.ConnectTimeout != nil {
		options.ConnectTimeout = *p.overrideDialer.ConnectTimeout
	}
	if p.overrideDialer.TCPFastOpen != nil {
		options.TCPFastOpen = *p.overrideDialer.TCPFastOpen
	}
	if p.overrideDialer.TCPMultiPath != nil {
		options.TCPMultiPath = *p.overrideDialer.TCPMultiPath
	}
	if p.overrideDialer.UDPFragment != nil {
		options.UDPFragment = p.overrideDialer.UDPFragment
	}
	if p.overrideDialer.DomainStrategy != nil {
		options.UDPFragment = p.overrideDialer.UDPFragment
	}
	if p.overrideDialer.FallbackDelay != nil {
		options.FallbackDelay = *p.overrideDialer.FallbackDelay
	}
	return options
}
