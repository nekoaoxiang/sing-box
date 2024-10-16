package provider

import (
	"github.com/sagernet/sing-box/option"
)

func newSMuxOptions(proxy map[string]any) *option.OutboundMultiplexOptions {
	if smux, exists := proxy["smux"].(map[string]any); exists {
		options := option.OutboundMultiplexOptions{}

		if enabled, exists := smux["enabled"].(bool); exists {
			options.Enabled = enabled
		}
		if protocol, exists := smux["protocol"].(string); exists {
			options.Protocol = protocol
		}
		if maxConnections, exists := smux["max-connections"].(int); exists {
			options.MaxConnections = maxConnections
		}
		if maxStreams, exists := smux["max-streams"].(int); exists {
			options.MaxStreams = maxStreams
		}
		if minStreams, exists := smux["min-streams"].(int); exists {
			options.MinStreams = minStreams
		}
		return &options
	}
	return nil
}
