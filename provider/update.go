package provider

import (
	"time"

	"github.com/sagernet/sing-box/adapter"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service"
)

func (l *Local) UpdateProvider() error {
	l.access.RLock()
	defer l.access.RUnlock()
	l.logger.DebugContext(l.ctx, "update start")

	err := l.parseProviderFile()
	if err != nil {
		return err
	}
	err = l.UpdateGroup()
	if err != nil {
		l.logger.ErrorContext(l.ctx, "update group failed: ", err)
		return nil
	}

	l.logger.InfoContext(l.ctx, "update success")
	return nil
}

func (r *Remote) UpdateProvider() error {
	r.access.RLock()
	defer r.access.RUnlock()
	r.logger.DebugContext(r.ctx, "update start")

	r.logger.DebugContext(r.ctx, "update: fetch Url")
	content, info, err := r.fetch()
	if err != nil {
		r.logger.ErrorContext(r.ctx, "update: fetch Url failed: ", err)
		return err
	}
	if content == nil {
		r.logger.InfoContext(r.ctx, "update success")
		return E.New("not modified")
	}

	r.logger.DebugContext(r.ctx, "update: parse raw info")
	subInfo, ok := parseRawInfo(info)
	if ok {
		r.subInfo = subInfo
	}

	r.logger.DebugContext(r.ctx, "update: parser raw config")
	options, err := NewParser(r.ctx, content)
	if err != nil {
		r.logger.ErrorContext(r.ctx, "update: parser raw config failed: ", err)
		return err
	}

	r.UpdateOutbounds(r.lastOutOpts, options.Outbounds)

	if r.path != "" {
		r.saveCacheContent(subInfo, options)
	}

	r.lastUpdateTime = time.Now()

	r.logger.DebugContext(r.ctx, "update: update groups")
	err = r.UpdateGroup()
	if err != nil {
		r.logger.ErrorContext(r.ctx, "update groups failed: ", err)
		return err
	}

	r.logger.InfoContext(r.ctx, "update success")
	return nil
}

func (p *MyProviderAdapter) UpdateGroup() error {
	var groups string
	outbound := service.FromContext[adapter.OutboundManager](p.ctx)
	for _, outbound := range outbound.Outbounds() {
		group, found := outbound.(adapter.OutboundGroup)
		if found {
			groups = groups + ` [` + group.Tag() + `]`
			err := group.UpdateGroup()
			if err != nil {
				return E.Cause(err, "outbound/", group.Type(), "[", group.Tag())
			}
		}
	}
	if groups != "" {
		p.logger.DebugContext(p.ctx, "update group", groups)
	}
	return nil
}
