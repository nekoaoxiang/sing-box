package provider

import (
	"context"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/taskmonitor"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/service"
)

func (p *myProviderAdapter) startOutbounds(outboundList []adapter.Outbound) (map[string]adapter.Outbound, []adapter.Outbound, error) {
	monitor := taskmonitor.New(p.logger, C.StartTimeout)
	started := make(map[string]bool)
	for _, outboundToStart := range outboundList {
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
				return nil, nil, E.Cause(err, "start outbound/", outboundToStart.Type(), "[", outboundTag, "]")
			}
		}
	}

	outbounds := make(map[string]adapter.Outbound)
	for _, out := range outbounds {
		tag := out.Tag()
		outbounds[tag] = out
	}
	return outbounds, outboundList, nil
}

func (p *myProviderAdapter) CreateOutbounds(ctx context.Context, router adapter.Router, options []option.Outbound) ([]adapter.Outbound, error) {
	outboundRegistry := service.FromContext[adapter.OutboundRegistry](ctx)

	outbounds := []adapter.Outbound{}

	for i, outboundOptions := range options {
		var currentOutbound adapter.Outbound
		var tag string
		var err error
		switch outboundOptions.Type {
		case C.TypeDirect, C.TypeBlock, C.TypeDNS, C.TypeSelector, C.TypeURLTest:
			continue
		default:
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
			currentOutbound, err = outboundRegistry.CreateOutbound(
				outboundCtx,
				router,
				p.logger,
				tag,
				outboundOptions.Type,
				outboundOptions.Options,
			)
			if err != nil {
				p.logger.WarnContext(ctx, "create outbound/", outboundOptions.Type, "[", outboundOptions.Tag, "]", " failed: ", err)
				continue
			}
			outbounds = append(outbounds, currentOutbound)
		}
	}
	if len(outbounds) > 0 && len(outbounds) == 0 && p.logger != nil {
		p.logger.WarnContext(ctx, "parse provider[", p.Adapter.Tag(), "] failed: missing valid outbound")
	}
	return outbounds, nil
}
