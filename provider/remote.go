package provider

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
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
			ctx:     ctx,
			logger:  logger,

			factory: factory,

			path: filemanager.BasePath(ctx, options.Path),
		},
		url:      options.Url,
		ua:       ua,
		interval: downloadInterval,
		detour:   options.Detour,
	}
	provider.provider = provider.newProviderManager()
	provider.endpoint = provider.newEndpointManager()
	provider.outbound = provider.newOutboundManager()

	process, err := NewProcessOptions(options.Filter)
	if err != nil {
		return nil, err
	}
	provider.process = process
	return provider, nil
}

func (r *Remote) Start(stage adapter.StartStage) error {
	switch stage {
	case adapter.StartStateInitialize:
		err := r.startManager()
		if err != nil {
			return err
		}
		err = r.parseCacheFile()
		if err != nil {
			return err
		}
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
	select {
	case <-time.After(initialWait): // 不要用 time.Sleep
	case <-p.ctx.Done():
		return
	}
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
	if r.ticker != nil {
		r.ticker.Stop()
	}
	var err error
	for _, closeItem := range []struct {
		name    string
		service adapter.Lifecycle
	}{
		{"endpoint", r.endpoint},
		{"outbound", r.outbound},
		{"provider", r.provider},
	} {
		r.logger.Trace("close ", closeItem.name)
		startTime := time.Now()
		err = E.Append(err, closeItem.service.Close(), func(err error) error {
			return E.Cause(err, "close ", closeItem.name)
		})
		r.logger.Trace("close ", closeItem.name, " completed (", F.Seconds(time.Since(startTime).Seconds()), "s)")
	}
	return nil
}
