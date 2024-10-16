package provider

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	"github.com/sagernet/sing-box/common/urltest"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/group"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

func RegisterRemote(registry *provider.Registry) {
	provider.Register[option.RemoteProviderOptions](registry, C.TypeRemote, NewRemote)
}

var (
	_ adapter.Provider = (*Remote)(nil)
)

type Remote struct {
	myProviderAdapter
	url      string
	ua       string
	interval time.Duration
	lastEtag string
	detour   string
	dialer   N.Dialer

	ticker *time.Ticker
}

func NewRemote(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.RemoteProviderOptions) (adapter.Provider, error) {
	if options.Url == "" {
		return nil, E.New("missing url")
	}

	var healthcheckInterval time.Duration
	var healthcheckUrl string
	var healthcheckEnable bool

	if options.HealthCheck != nil {
		healthcheckInterval = time.Duration(options.HealthCheck.Interval)
		healthcheckUrl = options.HealthCheck.Url
		healthcheckEnable = options.HealthCheck.Enable
	}

	if healthcheckInterval == 0 {
		healthcheckInterval = C.DefaultURLTestInterval
	}
	if healthcheckUrl == "" {
		healthcheckUrl = "https://www.gstatic.com/generate_204"
	}
	if !healthcheckEnable {
		healthcheckEnable = false
	}

	parsedURL, err := url.Parse(options.Url)
	if err != nil {
		return nil, err
	}
	switch parsedURL.Scheme {
	case "":
		parsedURL.Scheme = "http"
	case "http", "https":
	default:
		return nil, E.New("invalid url scheme")
	}

	ua := options.UserAgent
	if ua == "" {
		ua = "sing-box " + C.Version + "; Clash compatible"
	}
	downloadInterval := time.Duration(options.Interval)
	if downloadInterval == 0 {
		downloadInterval = 1 * time.Hour
	}

	ctx, cancel := context.WithCancel(ctx)

	var healthcheckHistory *urltest.HistoryStorage
	if healthcheckHistory = service.PtrFromContext[urltest.HistoryStorage](ctx); healthcheckHistory != nil {
	} else if clashServer := service.FromContext[adapter.ClashServer](ctx); clashServer != nil {
		healthcheckHistory = clashServer.HistoryStorage()
	} else {
		healthcheckHistory = urltest.NewHistoryStorage()
	}

	provider := &Remote{
		myProviderAdapter: myProviderAdapter{
			Adapter: provider.NewAdapter(C.TypeRemote, tag),
			ctx:     ctx,
			cancel:  cancel,
			logger:  logger,
			router:  router,

			outboundManager: service.FromContext[adapter.OutboundManager](ctx),

			path:         options.Path,
			pauseManager: service.FromContext[pause.Manager](ctx),

			HealthCheck: HealthCheck{
				healthCheckEnable:   healthcheckEnable,
				healthCheckUrl:      healthcheckUrl,
				healthCheckInterval: healthcheckInterval,
				healchcheckHistory:  healthcheckHistory,
			},
		},
		url:      parsedURL.String(),
		ua:       ua,
		interval: downloadInterval,
		detour:   options.Detour,
	}
	if options.Filter != nil {
		if options.Filter.Includes != nil {
			includes, err := group.NewProviderFilter(options.Filter.Includes)
			if err != nil {
				return nil, err
			}
			provider.includes = includes
		}

		if options.Filter.Excludes != nil {
			excludes, err := group.NewProviderFilter(options.Filter.Excludes)
			if err != nil {
				return nil, err
			}
			provider.excludes = excludes
		}
	}

	outboundList, subInfo, lastUpdatedTime, err := provider.parseCacheFile()

	if err != nil {
		return provider, nil
	}
	provider.lastUpdateTime = lastUpdatedTime
	provider.subInfo = subInfo

	if outboundList == nil {
		return provider, nil
	}

	outbounds := make(map[string]adapter.Outbound)
	for _, out := range outboundList {
		outbounds[out.Tag()] = out
	}

	provider.outbounds = outbounds
	provider.outboundList = outboundList

	return provider, nil
}

func (p *Remote) PostStart() error {
	var dialer N.Dialer
	if p.detour != "" {
		outbound, loaded := p.outboundManager.Outbound(p.detour)
		if !loaded {
			return E.New("download_detour not found: ", p.detour)
		}
		dialer = outbound
	} else {
		dialer = p.outboundManager.Default()
	}
	p.dialer = dialer
	go p.loopUpdate()
	go p.loopHealthCheck()
	return nil
}

func (p *Remote) loopUpdate() {
	timeSinceLastUpdate := time.Since(p.lastUpdateTime)
	initialWait := p.interval - timeSinceLastUpdate
	if initialWait < 0 {
		initialWait = 0
	}
	time.Sleep(initialWait)
	p.UpdateProvider(p.ctx)

	p.ticker = time.NewTicker(p.interval)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.ticker.C:
			p.pauseManager.WaitActive()
			p.UpdateProvider(p.ctx)
		}
	}
}

func (p *Remote) fetchAndParseURL(ctx context.Context) ([]byte, string, error) {
	defer runtime.GC()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: C.TCPTimeout,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return p.dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
			},
		},
	}

	request, err := http.NewRequestWithContext(ctx, "GET", p.url, nil)
	if err != nil {
		return nil, "", E.New("failed to create request: ", err)
	}

	if p.lastEtag != "" {
		request.Header.Set("If-None-Match", p.lastEtag)
	}
	request.Header.Set("User-Agent", p.ua)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, "", E.New("failed to execute request: ", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		if eTagHeader := response.Header.Get("Etag"); eTagHeader != "" {
			p.lastEtag = eTagHeader
		}
	case http.StatusNotModified:
		p.logger.InfoContext(ctx, "update provider[", p.Adapter.Tag(), "] not modified")
		return nil, "", nil
	default:
		return nil, "", E.New("unexpected status: ", response.Status)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", E.New("failed to read response body: ", err)
	}
	if len(content) == 0 {
		return nil, "", E.New("empty response")
	}

	info := response.Header.Get("subscription-userinfo")

	return content, info, nil
}

func (p *Remote) Close() error {
	if p.ticker != nil {
		p.ticker.Stop()
	}
	return p.myProviderAdapter.Close()
}
