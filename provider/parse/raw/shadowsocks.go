package provider

import (
	"net/url"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

func newRawShadowsocks(link *url.URL) (*option.Outbound, error) {
	if link.User == nil || link.User.Username() == "" {
		return nil, E.New("missing user info")
	}
	var options option.ShadowsocksOutboundOptions
	options.ServerOptions.Server = link.Hostname()
	options.ServerOptions.ServerPort = StringToType[uint16](link.Port())
	password, _ := link.User.Password()
	if password == "" {
		return nil, E.New("bad user info")
	}
	options.Method = link.User.Username()
	options.Password = password
	plugin := link.Query().Get("plugin")
	options.Plugin, options.PluginOptions = shadowsocksPlugin(plugin)

	outbound := &option.Outbound{
		Type: C.TypeShadowsocks,
		Tag:  link.Fragment,
	}
	outbound.Options = &options
	return outbound, nil
}

// shadowsocksPlugin splits plugin string into name and option part.
func shadowsocksPlugin(plugin string) (name, opts string) {
	parts := strings.SplitN(plugin, ";", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
