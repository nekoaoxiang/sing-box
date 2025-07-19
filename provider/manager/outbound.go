package manager

import (
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/service"
)

func (o *Manager) newOutboundManager() {
	outboundRegistry := service.FromContext[adapter.OutboundRegistry](o.ctx)
	NewOutbound := outbound.NewManager(o.logger, outboundRegistry, o.endpoint, o.provider, "")
	common.Close(o.outbound)
	o.outbound = NewOutbound
}

func (p *Manager) newOutbounds(options []option.Outbound) error {
	p.logger.DebugContext(p.ctx, "create outbounds")
	err := p.createOutbounds(options)
	if err != nil {
		return E.Cause(err, "create outbounds:")
	}

	p.logger.DebugContext(p.ctx, "start outbounds")
	err = p.start(p.outbound)
	if err != nil {
		return E.Cause(err, "start outbounds:")
	}
	return nil
}

func (p *Manager) createOutbounds(options []option.Outbound) error {
	for i, outboundOptions := range options {
		var tag string
		if outboundOptions.Tag != "" {
			tag = outboundOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		outboundCtx := p.ctx
		if tag != "" {
			// TODO: remove this
			outboundCtx = adapter.WithContext(outboundCtx, &adapter.InboundContext{
				Outbound: tag,
			})
		}
		err := p.outbound.Create(
			outboundCtx,
			p.router,
			p.factory.NewLogger(F.ToString("provider[", p.tag, "]/outbound[", tag, "]")),
			tag,
			outboundOptions.Type,
			outboundOptions.Options,
		)
		if err != nil {
			return E.Cause(err, "initialize outbound[", tag, "]")
		}
	}
	return nil
}
