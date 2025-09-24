package group

import (
	"context"
	"hash/maphash"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/interrupt"
	"github.com/sagernet/sing-box/common/urltest"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/provider"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/batch"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

func RegisterBalancer(registry *outbound.Registry) {
	outbound.Register[option.BalancerOutboundOptions](registry, C.TypeBalancer, NewLoadBalance)
}

var _ adapter.OutboundGroup = (*LoadBalance)(nil)

type LoadBalance struct {
	myGroupAdapter
	outbound.Adapter
	ctx        context.Context
	router     adapter.Router
	connection adapter.ConnectionManager
	logger     log.ContextLogger

	link        string
	interval    time.Duration
	idleTimeout time.Duration
	strategy    string

	group *LoadBalanceGroup
}

func NewLoadBalance(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.BalancerOutboundOptions) (adapter.Outbound, error) {
	outbound := &LoadBalance{
		myGroupAdapter: myGroupAdapter{
			uses:                options.Providers,
			includeAllProviders: options.IncludeAllProviders,
			providerManager:     service.FromContext[adapter.ProviderManager](ctx),
			outbound:            service.FromContext[adapter.OutboundManager](ctx),
			defaultTags:         options.Outbounds,
		},
		Adapter:    outbound.NewAdapter(C.TypeBalancer, tag, []string{N.NetworkTCP, N.NetworkUDP}, options.Outbounds),
		ctx:        ctx,
		router:     router,
		connection: service.FromContext[adapter.ConnectionManager](ctx),
		logger:     logger,

		link:        options.URL,
		interval:    time.Duration(options.Interval),
		idleTimeout: time.Duration(options.IdleTimeout),
		strategy:    options.Strategy,
	}
	if len(outbound.defaultTags) == 0 && len(outbound.uses) == 0 && !outbound.includeAllProviders {
		return nil, E.New("missing tags and uses")
	}
	process, err := provider.NewProcessOptions(options.Filter)
	if err != nil {
		return nil, err
	}
	outbound.process = process
	return outbound, nil
}

func (s *LoadBalance) UpdateGroup() error {
	err := s.getOutbounds()
	if err != nil {
		return E.New("update group outbound failed: ", s.Tag())
	}

	outbounds := []adapter.Outbound{}
	for _, tag := range s.tags {
		detour := s.outbounds[tag]
		outbounds = append(outbounds, detour)
	}

	s.group.updateGroup(outbounds)

	return nil
}

func (s *LoadBalance) Start() error {
	err := s.getOutbounds()
	if err != nil {
		return err
	}

	outbounds := []adapter.Outbound{}
	for _, tag := range s.tags {
		detour := s.outbounds[tag]
		outbounds = append(outbounds, detour)
	}
	group, err := NewBalancerGroup(s.ctx, s.outbound, s.logger, outbounds, s.link, s.interval, s.idleTimeout, s.strategy)
	if err != nil {
		return err
	}
	s.group = group
	return nil
}

func (s *LoadBalance) PostStart() error {
	s.group.PostStart()
	return nil
}

func (s *LoadBalance) Close() error {
	return common.Close(
		common.PtrOrNil(s.group),
	)
}

func (s *LoadBalance) All() []string {
	return s.tags
}

func (s *LoadBalance) Now() string {
	return s.group.selectedOutbound.Tag()
}

func (s *LoadBalance) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	s.group.Touch()
	outbound, _ := s.group.Select(network, destination)
	if outbound == nil {
		return nil, E.New("missing supported outbound")
	}
	conn, err := outbound.DialContext(ctx, network, destination)
	if err == nil {
		return s.group.interruptGroup.NewConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}
	s.logger.ErrorContext(ctx, err)
	s.group.history.DeleteURLTestHistory(outbound.Tag())
	return nil, err
}

func (s *LoadBalance) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	s.group.Touch()
	var outbound adapter.Outbound
	outbound, _ = s.group.Select(N.NetworkUDP, destination)
	if outbound == nil {
		return nil, E.New("missing supported outbound")
	}
	conn, err := outbound.ListenPacket(ctx, destination)
	if err == nil {
		return s.group.interruptGroup.NewPacketConn(conn, interrupt.IsExternalConnectionFromContext(ctx)), nil
	}
	s.logger.ErrorContext(ctx, err)
	s.group.history.DeleteURLTestHistory(outbound.Tag())
	return nil, err
}

func (s *LoadBalance) NewConnectionEx(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, onClose N.CloseHandlerFunc) {
	ctx = interrupt.ContextWithIsExternalConnection(ctx)
	s.connection.NewConnection(ctx, s, conn, metadata, onClose)
}

func (s *LoadBalance) NewPacketConnectionEx(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext, onClose N.CloseHandlerFunc) {
	ctx = interrupt.ContextWithIsExternalConnection(ctx)
	s.connection.NewPacketConnection(ctx, s, conn, metadata, onClose)
}

type LoadBalanceGroup struct {
	ctx              context.Context
	outbound         adapter.OutboundManager
	pause            pause.Manager
	pauseCallback    *list.Element[pause.Callback]
	logger           log.Logger
	outbounds        []adapter.Outbound
	link             string
	interval         time.Duration
	idleTimeout      time.Duration
	history          adapter.URLTestHistoryStorage
	checking         atomic.Bool
	selectedOutbound adapter.Outbound
	interruptGroup   *interrupt.Group
	access           sync.Mutex
	ticker           *time.Ticker
	close            chan struct{}
	started          bool
	lastActive       common.TypedValue[time.Time]
	strategy         strategyFunc
}

func NewBalancerGroup(ctx context.Context, outboundManager adapter.OutboundManager, logger log.Logger, outbounds []adapter.Outbound, link string, interval time.Duration, idleTimeout time.Duration, strategy string) (*LoadBalanceGroup, error) {
	if interval == 0 {
		interval = C.DefaultURLTestInterval
	}
	if idleTimeout == 0 {
		idleTimeout = C.DefaultURLTestIdleTimeout
	}
	if interval > idleTimeout {
		return nil, E.New("interval must be less or equal than idle_timeout")
	}
	if strategy == "" {
		strategy = "round-robin"
	}

	var strategyFunc strategyFunc
	switch strategy {
	case "consistent-hashing":
		strategyFunc = strategyConsistentHashing()
	case "round-robin":
		strategyFunc = strategyRoundRobin()
	default:
		return nil, E.New("unsupported strategy: ", strategy)
	}
	var history adapter.URLTestHistoryStorage
	if historyFromCtx := service.PtrFromContext[urltest.HistoryStorage](ctx); historyFromCtx != nil {
		history = historyFromCtx
	} else if clashServer := service.FromContext[adapter.ClashServer](ctx); clashServer != nil {
		history = clashServer.HistoryStorage()
	} else {
		history = urltest.NewHistoryStorage()
	}
	return &LoadBalanceGroup{
		ctx:            ctx,
		outbound:       outboundManager,
		logger:         logger,
		outbounds:      outbounds,
		link:           link,
		interval:       interval,
		idleTimeout:    idleTimeout,
		strategy:       strategyFunc,
		history:        history,
		close:          make(chan struct{}),
		pause:          service.FromContext[pause.Manager](ctx),
		interruptGroup: interrupt.NewGroup(),
	}, nil
}

func (g *LoadBalanceGroup) Select(network string, destination M.Socksaddr) (adapter.Outbound, bool) {
	var outbounds []adapter.Outbound
	for _, detour := range g.outbounds {
		if !common.Contains(detour.Network(), network) {
			continue
		}
		outbounds = append(outbounds, detour)
	}
	detour := g.strategy(outbounds, destination, true)
	if detour == nil {
		g.logger.Warn("no available outbounds for network: ", network)
		for _, detour := range g.outbounds {
			if !common.Contains(detour.Network(), network) {
				continue
			}
			return detour, false
		}
		return nil, false
	}
	g.selectedOutbound = detour
	return detour, false
}

func (g *LoadBalanceGroup) Touch() {
	if !g.started {
		return
	}
	g.access.Lock()
	defer g.access.Unlock()
	if g.ticker != nil {
		g.lastActive.Store(time.Now())
		return
	}
	g.ticker = time.NewTicker(g.interval)
	go g.loopCheck()
	g.pauseCallback = pause.RegisterTicker(g.pause, g.ticker, g.interval, nil)
}

func (g *LoadBalanceGroup) loopCheck() {
	if time.Since(g.lastActive.Load()) > g.interval {
		g.lastActive.Store(time.Now())
	}
	for {
		select {
		case <-g.close:
			return
		case <-g.ticker.C:
		}
		if time.Since(g.lastActive.Load()) > g.idleTimeout {
			g.access.Lock()
			g.ticker.Stop()
			g.ticker = nil
			g.pause.UnregisterCallback(g.pauseCallback)
			g.pauseCallback = nil
			g.access.Unlock()
			return
		}
	}
}

func (g *LoadBalanceGroup) PostStart() {
	g.access.Lock()
	defer g.access.Unlock()
	g.started = true
	g.lastActive.Store(time.Now())
	go g.CheckOutbounds(false)
}

func (g *LoadBalanceGroup) CheckOutbounds(force bool) {
	_, _ = g.urlTest(g.ctx, force)
}

func (g *LoadBalanceGroup) URLTest(ctx context.Context) (map[string]uint16, error) {
	return g.urlTest(ctx, false)
}

func (g *LoadBalanceGroup) urlTest(ctx context.Context, force bool) (map[string]uint16, error) {
	result := make(map[string]uint16)
	if g.checking.Swap(true) {
		return result, nil
	}
	defer g.checking.Store(false)
	b, _ := batch.New(ctx, batch.WithConcurrencyNum[any](10))
	checked := make(map[string]bool)
	var resultAccess sync.Mutex
	for _, detour := range g.outbounds {
		tag := detour.Tag()
		realTag := RealTag(detour)
		if checked[realTag] {
			continue
		}
		history := g.history.LoadURLTestHistory(realTag)
		if !force && history != nil && time.Since(history.Time) < g.interval {
			continue
		}
		checked[realTag] = true
		// p, loaded := g.outbound.Outbound(realTag)
		// if !loaded {
		// 	continue
		// }
		b.Go(realTag, func() (any, error) {
			testCtx, cancel := context.WithTimeout(g.ctx, C.TCPTimeout)
			defer cancel()
			t, err := urltest.URLTest(testCtx, g.link, detour)
			if err != nil {
				g.logger.Debug("outbound ", tag, " unavailable: ", err)
				g.history.DeleteURLTestHistory(realTag)
			} else {
				g.logger.Debug("outbound ", tag, " available: ", t, "ms")
				g.history.StoreURLTestHistory(realTag, &adapter.URLTestHistory{
					Time:  time.Now(),
					Delay: t,
				})
				resultAccess.Lock()
				result[tag] = t
				resultAccess.Unlock()
			}
			return nil, nil
		})
	}
	b.Wait()
	return result, nil
}

func (s *LoadBalanceGroup) updateGroup(outbounds []adapter.Outbound) {
	s.outbounds = outbounds
}

type strategyFunc = func(outbounds []adapter.Outbound, destination M.Socksaddr, touch bool) adapter.Outbound

func strategyRoundRobin() strategyFunc {
	idx := 0
	idxMutex := sync.Mutex{}
	return func(outbounds []adapter.Outbound, destination M.Socksaddr, touch bool) adapter.Outbound {
		idxMutex.Lock()
		defer idxMutex.Unlock()
		length := len(outbounds)
		id := idx
		if touch {
			idx = (idx + 1) % length
		}
		return outbounds[id]
	}
}

func strategyConsistentHashing() strategyFunc {
	var globalSeed = maphash.MakeSeed()
	return func(outbounds []adapter.Outbound, destination M.Socksaddr, touch bool) adapter.Outbound {
		length := len(outbounds)
		if length == 0 {
			return nil
		}
		if length == 1 {
			return outbounds[0]
		}
		var target string
		if destination.Fqdn != "" {
			if etld, err := publicsuffix.EffectiveTLDPlusOne(destination.Fqdn); err == nil {
				target = etld
			} else {
				target = destination.Fqdn
			}
		} else {
			target = destination.String()
		}
		key := maphash.String(globalSeed, target)
		idx := jumpHash(key, int32(length))
		return outbounds[idx]
	}
}

func jumpHash(key uint64, buckets int32) int32 {
	var b, j int64
	for j < int64(buckets) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(1<<31) / float64((key>>33)+1)))
	}
	return int32(b)
}
