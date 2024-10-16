package provider

import (
	"context"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/taskmonitor"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
)

func (p *myProviderAdapter) CreateOutbounds(ctx context.Context, router adapter.Router, options []option.Outbound) error {
	for i, outboundOptions := range options {
		var tag string
		if outboundOptions.Tag != "" {
			tag = outboundOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		outboundCtx := ctx
		if tag != "" {
			// TODO: remove this
			outboundCtx = adapter.WithContext(outboundCtx, &adapter.InboundContext{
				Outbound: tag,
			})
		}
		err := p.outboundManager.Create(
			outboundCtx,
			router,
			p.logger,
			tag,
			outboundOptions.Type,
			outboundOptions.Options,
		)
		if err != nil {
			return E.Cause(err, "initialize outbound[", i, "]")
		}
	}
	return nil
}

func (p *myProviderAdapter) startOutbounds() error {
	monitor := taskmonitor.New(p.logger, C.StartTimeout)
	started := make(map[string]bool)
	for _, outboundToStart := range p.outboundManager.Outbounds() {
		outboundTag := outboundToStart.Tag()
		if started[outboundTag] {
			continue
		}
		started[outboundTag] = true
		if starter, isStarter := outboundToStart.(interface {
			Start() error
		}); isStarter {
			monitor.Start("start outbound/", outboundToStart.Type(), "[", outboundTag, "]")
			err := starter.Start()
			monitor.Finish()
			if err != nil {
				return E.Cause(err, "start outbound/", outboundToStart.Type(), "[", outboundTag, "]")
			}
		}
	}
	return nil
}
