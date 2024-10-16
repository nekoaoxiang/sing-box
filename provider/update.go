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
	lastUpdatedTime, err := p.parseProviderFile()
	if err != nil {
		return err
	}

	err = p.startOutbounds()
	if err != nil {
		return err
	}

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
	p.logger.DebugContext(ctx, "update: fetch and parse Url")
	rawContent, rawInfo, err := p.fetchAndParseURL(ctx)
	if err != nil {
		p.logger.ErrorContext(ctx, E.New("update provider", "[", p.Adapter.Tag(), "]", " failed ", err))
		return err
	}

	p.logger.DebugContext(ctx, "update: parser raw config")
	options, err := p.newParser(rawContent)
	if err != nil {
		return err
	}
	p.logger.DebugContext(ctx, "update: create outbounds")
	err = p.CreateOutbounds(p.ctx, p.router, options)
	if err != nil {
		return err
	}

	p.logger.DebugContext(ctx, "update: start outbounds")
	err = p.startOutbounds()
	if err != nil {
		return err
	}

	p.logger.DebugContext(ctx, "update: parse raw provider info")
	subInfo, ok := parseRawInfo(rawInfo)
	if ok {
		p.subInfo = subInfo
	}

	p.lastUpdateTime = time.Now()

	err = p.UpdateGroup()
	if err != nil {
		p.logger.ErrorContext(ctx, E.New("update outbound group failed.", err))
	}

	if p.path != "" {
		p.saveCacheContent(subInfo, options)
	}

	p.logger.InfoContext(ctx, "update success")

	return err
}

func (p *myProviderAdapter) UpdateGroup() error {
	for _, outbound := range p.defaultOutbound.Outbounds() {
		group, found := outbound.(adapter.OutboundGroup)
		if found {
			p.logger.Debug("update outbound/", group.Type(), "[", group.Tag())
			err := group.UpdateGroup(p.Adapter.Tag())
			if err != nil {
				return E.Cause(err, "update outbound/", group.Type(), "[", group.Tag())
			}
		}
	}
	return nil
}
