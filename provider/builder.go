package provider

import (
	"context"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/adapter/provider"
	"github.com/sagernet/sing-box/common/urltest"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing/service"
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
	defaultOutbound adapter.OutboundManager

	path string

	includes *group.FilterList
	excludes *group.FilterList

	lastUpdateTime time.Time
	pauseManager   pause.Manager
	checking       atomic.Bool
	updating       atomic.Bool

	HealthCheck

	subInfo *option.SubInfo
}

func NewOutboundManger(ctx context.Context, logger log.ContextLogger) *outbound.Manager {
	outboundRegistry := service.FromContext[adapter.OutboundRegistry](ctx)
	outboundManager := outbound.NewManager(logger, outboundRegistry, "")
	return outboundManager
}

func parseRawInfo(rawInfo string) (*option.SubInfo, bool) {
	var info = &option.SubInfo{}
	reg := regexp.MustCompile(`upload=[+-]?(\d*);[ \t]*download=[+-]?(\d*);[ \t]*total=[+-]?(\d*);[ \t]*expire=[+-]?(\d*)`)
	result := reg.FindStringSubmatch(rawInfo)

	if len(result) > 0 {
		upload, _ := strconv.ParseInt(result[1], 10, 64)
		download, _ := strconv.ParseInt(result[2], 10, 64)
		total, _ := strconv.ParseInt(result[3], 10, 64)
		expire, _ := strconv.ParseInt(result[4], 10, 64)

		info.Upload = upload
		info.Download = download
		info.Total = total
		info.Expire = expire
		return info, true
	}
	return info, false
}

func (p *myProviderAdapter) SubInfo() map[string]int64 {
	info := make(map[string]int64)
	if p.subInfo != nil {
		info["Upload"] = p.subInfo.Upload
		info["Download"] = p.subInfo.Download
		info["Total"] = p.subInfo.Total
		info["Expire"] = p.subInfo.Expire
	}
	return info
}

func (p *myProviderAdapter) LastUpdateTime() time.Time {
	return p.lastUpdateTime
}

func (p *myProviderAdapter) OutboundManager() adapter.OutboundManager {
	return p.outboundManager
}

func (p *myProviderAdapter) Close() error {
	if p.healthCheckTicker != nil {
		p.healthCheckTicker.Stop()
	}
	p.cancel()
	return nil
}
