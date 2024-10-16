package provider

import (
	"context"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
	E "github.com/sagernet/sing/common/exceptions"
)

func (p *Local) UpdateProvider(ctx context.Context) error {
	if p.updating.Swap(true) {
		return E.New("provider is updating")
	}
	defer p.updating.Store(false)
	ctx = log.ContextWithNewID(ctx)
	p.logger.DebugContext(ctx, "update provider/", p.Adapter.Type(), "[", p.Adapter.Tag(), "]")
	outbounds_adapter, lastUpdatedTime, err := p.parseProviderFile()
	if err != nil {
		return err
	}

	outbounds, outboundList, err := p.startOutbounds(outbounds_adapter)
	if err != nil {
		return err
	}

	p.outbounds = outbounds
	p.outboundList = outboundList
	p.lastUpdateTime = lastUpdatedTime

	err = p.UpdateGroup()
	if err != nil {
		p.logger.ErrorContext(ctx, E.New("update provider", "[", p.Adapter.Tag(), "]", " outbound failed.", err))
	}

	return err
}

func (p *Remote) UpdateProvider(ctx context.Context) error {
	if p.updating.Swap(true) {
		return E.New("provider is updating")
	}
	defer p.updating.Store(false)
	ctx = log.ContextWithNewID(ctx)
	p.logger.DebugContext(ctx, "update provider/", p.Adapter.Type(), "[", p.Adapter.Tag(), "]")
	rawContent, rawInfo, err := p.fetchAndParseURL(ctx)
	if err != nil {
		p.logger.ErrorContext(ctx, E.New("update provider", "[", p.Adapter.Tag(), "]", " failed ", err))
		return err
	}

	outbounds_option, err := p.newParser(rawContent)
	if err != nil {
		return err
	}
	outbounds_adapter, err := p.CreateOutbounds(p.ctx, p.router, outbounds_option)
	if err != nil {
		return err
	}

	outbounds, outboundList, err := p.startOutbounds(outbounds_adapter)
	if err != nil {
		return err
	}

	subInfo, ok := parseRawInfo(rawInfo)
	if ok {
		p.subInfo = subInfo
	}

	p.outbounds = outbounds
	p.outboundList = outboundList

	p.lastUpdateTime = time.Now()

	err = p.UpdateGroup()
	if err != nil {
		p.logger.ErrorContext(ctx, E.New("update outbound group failed.", err))
	}

	if p.path != "" {
		p.saveCacheContent(subInfo, outbounds_option)
	}

	p.logger.InfoContext(ctx, "update success")

	return err
}

func (p *myProviderAdapter) UpdateGroup() error {
	for _, outbound := range p.outboundManager.Outbounds() {
		if group, ok := outbound.(adapter.OutboundGroup); ok {
			p.logger.Debug("update outbound/", group.Type(), "[", group.Tag())
			err := group.UpdateGroup(p.Adapter.Tag())
			if err != nil {
				return E.Cause(err, "update outbound/", group.Type(), "[", group.Tag())
			}
		}
	}
	return nil
}
