package provider

import (
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

type AnyTLSOption struct {
	BaseProxy                `yaml:",inline"`
	DialerOptions            `yaml:",inline"`
	Password                 string `yaml:"password"`
	*TLSOption               `yaml:",inline"`
	IdleSessionCheckInterval int `yaml:"idle-session-check-interval,omitempty"`
	IdleSessionTimeout       int `yaml:"idle-session-timeout,omitempty"`
	MinIdleSession           int `yaml:"min-idle-session,omitempty"`
}

func newClashAnyTLS(tag string, proxy AnyTLSOption) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeAnyTLS,
		Tag:  tag,
	}
	Options := &option.AnyTLSOutboundOptions{
		ServerOptions: option.ServerOptions{
			Server:     proxy.BaseProxy.Server,
			ServerPort: proxy.BaseProxy.Port,
		},
		Password:                 proxy.Password,
		IdleSessionCheckInterval: badoption.Duration(proxy.IdleSessionCheckInterval),
		IdleSessionTimeout:       badoption.Duration(proxy.IdleSessionTimeout),
		MinIdleSession:           proxy.MinIdleSession,

		DialerOptions: newDialerOptions(proxy.DialerOptions),
	}

	Options.TLS = newTLSOptions(proxy.TLSOption)

	outbound.Options = Options
	return outbound, nil
}
