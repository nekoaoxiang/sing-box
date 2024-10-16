package provider

import (
	"context"
	"sync"
	"time"

	"github.com/sagernet/sing-box/common/urltest"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing/common/batch"
)

func (p *myProviderAdapter) loopHealthCheck() {
	if !p.healthCheckEnable {
		return
	}
	ctx, cancel := context.WithCancel(p.ctx)
	defer cancel()
	p.healthCheckTicker = time.NewTicker(p.healthCheckInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.healthCheckTicker.C:
			p.pauseManager.WaitActive()
			p.Healthcheck(p.ctx, p.healthCheckUrl)
		}
	}
}

func (p *myProviderAdapter) Healthcheck(ctx context.Context, link string) map[string]uint16 {
	result := make(map[string]uint16)
	if p.checking.Swap(true) {
		return result
	}
	defer p.checking.Store(false)
	b, _ := batch.New(ctx, batch.WithConcurrencyNum[any](10))
	checked := make(map[string]bool)
	var resultAccess sync.Mutex
	for _, detour := range p.outboundManager.Outbounds() {
		tag := detour.Tag()
		if checked[tag] {
			continue
		}
		checked[tag] = true
		detour, loaded := p.outboundManager.Outbound(tag)
		if !loaded {
			continue
		}
		b.Go(tag, func() (any, error) {
			ctx, cancel := context.WithTimeout(log.ContextWithNewID(context.Background()), C.TCPTimeout)
			defer cancel()
			t, err := urltest.URLTest(ctx, link, detour)
			if err != nil {
				p.logger.DebugContext(ctx, "outbound ", tag, " unavailable: ", err)
				p.healchcheckHistory.DeleteURLTestHistory(tag)
			} else {
				p.logger.DebugContext(ctx, "outbound ", tag, " available: ", t, "ms")
				p.healchcheckHistory.StoreURLTestHistory(tag, &urltest.History{
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
	return result
}
