package provider

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing-box/provider/manager"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/ntp"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/filemanager"
)

func RegisterRemote(registry *provider.Registry) {
	provider.Register[option.RemoteProviderOptions](registry, C.TypeRemote, NewRemote)
}

var (
	_ adapter.Provider = (*Remote)(nil)
)

type Remote struct {
	MyProviderAdapter
	url      string
	ua       string
	interval time.Duration
	detour   string

	lastEtag string
	dialer   N.Dialer
	ticker   *time.Ticker
}

func NewRemote(ctx context.Context, router adapter.Router, factory log.Factory, tag string, options option.RemoteProviderOptions) (adapter.Provider, error) {
	if options.Url == "" {
		return nil, E.New("missing url")
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

	logger := factory.NewLogger(F.ToString("provider/", options.Type, "[", tag, "]"))

	provider := &Remote{
		MyProviderAdapter: MyProviderAdapter{
			Adapter: provider.NewAdapter(C.TypeRemote, tag),
			Manager: manager.NewManager(ctx, logger, router, tag, factory),
			ctx:     ctx,
			logger:  logger,

			path: filemanager.BasePath(ctx, options.Path),
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
	return provider, nil
}

func (r *Remote) Start(stage adapter.StartStage) error {
	switch stage {
	case adapter.StartStateInitialize:
		r.parseCacheFile()
	case adapter.StartStatePostStart:
		outbound := service.FromContext[adapter.OutboundManager](r.ctx)
		var dialer N.Dialer
		if r.detour != "" {
			outbound, loaded := outbound.Outbound(r.detour)
			if !loaded {
				return E.New("download_detour not found: ", r.detour)
			}
			dialer = outbound
		} else {
			dialer = outbound.Default()
		}
		r.dialer = dialer

		go r.loopUpdate()
	}
	return nil
}

func (p *Remote) loopUpdate() {
	timeSinceLastUpdate := time.Since(p.lastUpdateTime)
	initialWait := max(p.interval-timeSinceLastUpdate, 0)
	time.Sleep(initialWait)
	p.UpdateProvider()
	p.ticker = time.NewTicker(p.interval)
	for {
		select {
		case <-p.ctx.Done():
			return
		case <-p.ticker.C:
			p.UpdateProvider()
		}
	}
}

func (r *Remote) fetch() ([]byte, string, error) {
	defer runtime.GC()
	httpClient := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2:   true,
			TLSHandshakeTimeout: C.TCPTimeout,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return r.dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
			},
			TLSClientConfig: &tls.Config{
				Time:    ntp.TimeFuncFromContext(r.ctx),
				RootCAs: adapter.RootPoolFromContext(r.ctx),
			},
		},
	}
	request, err := http.NewRequest("GET", r.url, nil)
	if err != nil {
		return nil, "", err
	}
	if r.lastEtag != "" {
		request.Header.Set("If-None-Match", r.lastEtag)
	}
	request.Header.Set("User-Agent", r.ua)
	response, err := httpClient.Do(request.WithContext(r.ctx))
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		if eTagHeader := response.Header.Get("Etag"); eTagHeader != "" {
			r.lastEtag = eTagHeader
		}
	case http.StatusNotModified:
		r.logger.DebugContext(r.ctx, "update: not modified")
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

func (r *Remote) Close() error {
	if r.ticker == nil {
		return nil
	}
	r.ticker.Stop()
	return nil
}
