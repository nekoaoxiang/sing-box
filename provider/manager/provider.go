package manager

import (
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/service"
)

func (p *Manager) newProviderManager() {
	providerRegistry := service.FromContext[adapter.ProviderRegistry](p.ctx)
	p.provider = provider.NewManager(p.logger, providerRegistry)
}

func (p *Manager) newProviders(options []option.Provider) error {
	p.logger.DebugContext(p.ctx, "create providers")
	err := p.createProviders(options)
	if err != nil {
		return err
	}
	return nil
}

func (p *Manager) createProviders(options []option.Provider) error {
	for i, providerOptions := range options {
		var tag string
		if providerOptions.Tag != "" {
			tag = providerOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		err := p.provider.Create(p.ctx,
			p.router,
			p.factory,
			tag+" by "+p.tag,
			providerOptions.Type,
			providerOptions.Options,
		)
		if err != nil {
			return E.Cause(err, "initialize provider[", i, "]")
		}
	}
	return nil
}
