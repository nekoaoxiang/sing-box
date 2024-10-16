package provider

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	"github.com/sagernet/sing-box/common/urltest"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing/service/pause"
)

type HealthCheck struct {
	healthCheckEnable   bool
	healthCheckUrl      string
	healthCheckInterval time.Duration

	healchcheckHistory *urltest.HistoryStorage
	healthCheckTicker  *time.Ticker
}

type myProviderAdapter struct {
	provider.Adapter

	ctx    context.Context
	cancel context.CancelFunc
	logger log.ContextLogger
	router adapter.Router

	outboundManager adapter.OutboundManager

	path string

	includes *group.FilterList
	excludes *group.FilterList

	outbounds    map[string]adapter.Outbound
	outboundList []adapter.Outbound

	lastUpdateTime time.Time
	pauseManager   pause.Manager
	checking       atomic.Bool
	updating       atomic.Bool

	HealthCheck

	subInfo *option.SubInfo
}

// type myProviderAdapter struct {
// 	ctx     context.Context
// 	cancel  context.CancelFunc
// 	router  adapter.Router
// 	logger  log.ContextLogger
// 	subInfo *option.SubInfo

// 	// Common config
// 	tag                 string
// 	path                string
// 	healthCheckEnable   bool
// 	healthCheckUrl      string
// 	healthCheckInterval time.Duration
// 	outboundOverride    *option.OutboundOverrideOptions
// 	healchcheckHistory  *urltest.HistoryStorage
// 	providerType        string
// 	lastUpdateTime      time.Time
// 	outbounds           []adapter.Outbound
// 	outboundByTag       map[string]adapter.Outbound
// 	includes            *O.FilterList
// 	excludes            *O.FilterList

// 	// Update cache
// 	checking     atomic.Bool
// 	updating     atomic.Bool
// 	pauseManager pause.Manager

// 	healthCheckTicker *time.Ticker
// 	close             chan struct{}
// }

// func New(ctx context.Context, router adapter.Router, logger log.ContextLogger, options option.Provider) (adapter.Provider, error) {
// 	switch options.Type {
// 	case "local":
// 		return NewLocalProvider(ctx, router, logger, options)
// 	case "remote":
// 		return NewRemoteProvider(ctx, router, logger, options)
// 	default:
// 		return nil, E.New("invalid provider type")
// 	}
// }

func (a *myProviderAdapter) SubInfo() map[string]int64 {
	info := make(map[string]int64)
	if a.subInfo != nil {
		info["Upload"] = a.subInfo.Upload
		info["Download"] = a.subInfo.Download
		info["Total"] = a.subInfo.Total
		info["Expire"] = a.subInfo.Expire
	}
	return info
}

func (a *myProviderAdapter) LastUpdateTime() time.Time {
	return a.lastUpdateTime
}

func (a *myProviderAdapter) Outbound(tag string) (adapter.Outbound, bool) {
	outbound, loaded := a.outbounds[tag]
	return outbound, loaded
}

func (p *myProviderAdapter) Outbounds() []adapter.Outbound {
	var outbounds []adapter.Outbound
	outbounds = append(outbounds, p.outboundList...)
	return outbounds
}

func (p *myProviderAdapter) Close() error {
	if p.healthCheckTicker != nil {
		p.healthCheckTicker.Stop()
	}
	// p.cancel()
	return nil
}
