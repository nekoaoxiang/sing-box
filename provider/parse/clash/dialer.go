package provider

import (
	"github.com/sagernet/sing-box/option"
)

type DialerOptions struct {
	IpVersion     string        `yaml:"ip-version,omitempty"`
	InterfaceName string        `yaml:"interface-name,omitempty"`
	RoutingMark   option.FwMark `yaml:"routing-mark,omitempty"`
	TFO           bool          `yaml:"tfo,omitempty"`
	MPTCP         bool          `yaml:"mptcp,omitempty"`
	DialerProxy   string        `yaml:"dialer-proxy,omitempty"`
}

func newDialerOptions(proxy DialerOptions) option.DialerOptions {
	options := option.DialerOptions{
		Detour:        proxy.DialerProxy,
		BindInterface: proxy.InterfaceName,
		RoutingMark:   option.FwMark(proxy.RoutingMark),
		TCPFastOpen:   proxy.TFO,
		TCPMultiPath:  proxy.MPTCP,
	}
	return options
}
