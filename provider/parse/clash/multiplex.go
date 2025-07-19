package provider

import (
	"github.com/sagernet/sing-box/option"
)

type smuxOpts struct {
	Enabled        bool           `yaml:"enabled"`
	Protocol       string         `yaml:"protocol"`
	MaxConnections int            `yaml:"max-connections"`
	MaxStreams     int            `yaml:"max-streams"`
	MinStreams     int            `yaml:"min-streams"`
	Padding        bool           `yaml:"padding"`
	Brutal         *BrutalOptions `yaml:"brutal-opts"`
}

type BrutalOptions struct {
	Enabled  bool `yaml:"enabled"`
	UpMbps   int  `yaml:"up"`
	DownMbps int  `yaml:"down"`
}

func newSMuxOptions(proxy *smuxOpts) *option.OutboundMultiplexOptions {
	if proxy != nil {
		options := &option.OutboundMultiplexOptions{
			Enabled:        proxy.Enabled,
			Protocol:       proxy.Protocol,
			MaxConnections: proxy.MaxConnections,
			MinStreams:     proxy.MaxStreams,
			MaxStreams:     proxy.MinStreams,
			Padding:        proxy.Padding,
			Brutal:         newBrutalOptions(proxy.Brutal),
		}
		return options
	}
	return nil
}

func newBrutalOptions(proxy *BrutalOptions) *option.BrutalOptions {
	options := &option.BrutalOptions{
		Enabled:  proxy.Enabled,
		UpMbps:   proxy.UpMbps,
		DownMbps: proxy.DownMbps,
	}
	return options
}
