package group

import (
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/provider"
	E "github.com/sagernet/sing/common/exceptions"
)

type myGroupAdapter struct {
	defaultTags []string
	outbound    adapter.OutboundManager

	tags      []string
	outbounds map[string]adapter.Outbound

	uses                []string
	includeAllProviders bool

	process *provider.ProcessOptions

	providerManager adapter.ProviderManager
}

func (g *myGroupAdapter) getOutbounds() error {
	outbounds := make(map[string]adapter.Outbound)
	tags := []string{}
	uses := []string{}
	providers := []adapter.Provider{}
	for i, tag := range g.defaultTags {
		detour, loaded := g.outbound.Outbound(tag)
		if !loaded {
			return E.New("outbound ", i, " not found: ", tag)
		}
		outbounds[tag] = detour
		tags = append(tags, tag)
	}

	if g.includeAllProviders {
		for _, provider := range g.providerManager.Providers() {
			g.uses = append(uses, provider.Tag())
		}
	}
	for i, use := range g.uses {
		provider, loaded := g.providerManager.Provider(use)
		if !loaded {
			return E.New("provider ", i, " not found: ", use)
		}
		providers = append(providers, provider)
	}
	for _, provider := range providers {
		if manager := provider.Outbound(); manager != nil {
			for _, outbound := range manager.Outbounds() {
				if !g.process.Process(outbound.Tag()) {
					continue
				}
				tag := outbound.Tag()
				outbounds[tag] = outbound
				tags = append(tags, tag)
			}
		}
	}
	if len(outbounds) == 0 {
		direct, _ := g.outbound.Outbound("direct")
		outbounds[direct.Tag()] = direct
		tags = append(tags, direct.Tag())
	}
	g.outbounds = outbounds
	g.tags = tags
	return nil
}
