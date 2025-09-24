package provider

import (
	"reflect"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/adapter/provider"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/direct"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/service"
)

func (p *MyProviderAdapter) newEndpointManager() adapter.EndpointManager {
	endpointRegistry := service.FromContext[adapter.EndpointRegistry](p.ctx)
	endpoint := endpoint.NewManager(p.logger, endpointRegistry)
	return endpoint
}

func (p *MyProviderAdapter) newProviderManager() adapter.ProviderManager {
	providerRegistry := service.FromContext[adapter.ProviderRegistry](p.ctx)
	provider := provider.NewManager(p.logger, providerRegistry)
	return provider
}

func (p *MyProviderAdapter) newOutboundManager() adapter.OutboundManager {
	outboundRegistry := service.FromContext[adapter.OutboundRegistry](p.ctx)
	outbound := outbound.NewManager(p.logger, outboundRegistry, p.endpoint, p.provider, "")
	outbound.Initialize(func() (adapter.Outbound, error) {
		return direct.NewOutbound(
			p.ctx,
			p.router,
			p.factory.NewLogger("outbound/direct"),
			"direct",
			option.DirectOutboundOptions{},
		)
	})
	return outbound
}

func (p *MyProviderAdapter) startManager() error {
	err := adapter.Start(p.logger, adapter.StartStateInitialize, p.outbound, p.endpoint, p.provider)
	if err != nil {
		return err
	}
	err = adapter.Start(p.logger, adapter.StartStateStart, p.outbound)
	if err != nil {
		return err
	}
	err = adapter.Start(p.logger, adapter.StartStatePostStart, p.outbound, p.endpoint, p.provider)
	if err != nil {
		return err
	}
	err = adapter.Start(p.logger, adapter.StartStateStarted, p.outbound, p.endpoint)
	if err != nil {
		return err
	}
	return nil
}

func (p *MyProviderAdapter) UpdateOutbounds(oldOpts []option.Outbound, newOpts []option.Outbound) {
	oldByTag := make(map[string]option.Outbound, len(oldOpts))
	for i, opt := range oldOpts {
		var tag string
		if opt.Tag != "" {
			tag = F.ToString(opt.Tag)
		} else {
			tag = F.ToString(i)
		}

		oldByTag[tag] = opt
	}

	needFullRebuild := false
	for i, opt := range newOpts {
		var tag string
		if opt.Tag != "" {
			tag = F.ToString(opt.Tag)
		} else {
			tag = F.ToString(i)
		}

		oldOpt, ok := oldByTag[tag]
		if !ok || !reflect.DeepEqual(opt, oldOpt) {
			needFullRebuild = true
			break
		}
	}

	if needFullRebuild {
		_ = p.createOutbound("direct", option.Outbound{
			Tag:  "direct",
			Type: "direct",
		})
		// 删除所有旧 outbound
		for _, outbound := range p.outbound.Outbounds() {
			tag := outbound.Tag()
			if tag == "direct" {
				continue
			}
			err := p.outbound.Remove(tag)
			if err != nil {
				p.logger.ErrorContext(p.ctx, "remove outbound[", tag, "] error: ", err)
			}
		}

		// 按 newOpts 顺序重新创建
		for i, opt := range newOpts {
			var tag string
			if opt.Tag != "" {
				tag = F.ToString(opt.Tag)
			} else {
				tag = F.ToString(i)
			}

			if !p.process.Process(tag) {
				continue
			}

			err := p.createOutbound(tag, opt)
			if err != nil {
				p.logger.ErrorContext(p.ctx, "create outbound[", tag, "] error: ", err)
			}
		}

		_ = p.outbound.Remove("direct")

		p.lastOutOpts = newOpts
	}
}

// createOutbound 抽象创建流程，便于统一日志与错误处理
func (p *MyProviderAdapter) createOutbound(tag string, opt option.Outbound) error {
	err := p.outbound.Create(
		adapter.WithContext(p.ctx, &adapter.InboundContext{
			Outbound: tag,
		}),
		p.router,
		p.factory.NewLogger(F.ToString("outbound/", opt.Type, "[", tag, "]")),
		tag,
		opt.Type,
		opt.Options,
	)
	return err
}
