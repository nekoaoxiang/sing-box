package provider

import (
	"context"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
)

type MyProviderAdapter struct {
	provider.Adapter
	// manager *manager.Manager

	ctx    context.Context
	logger log.ContextLogger
	access sync.RWMutex

	path string

	process *ProcessOptions

	lastUpdateTime time.Time
	subInfo        map[string]int64

	factory     log.Factory
	router      adapter.Router
	lastOutOpts []option.Outbound

	provider adapter.ProviderManager
	outbound adapter.OutboundManager
	endpoint adapter.EndpointManager
}

var rawInfoRegexp = regexp.MustCompile(`upload=[+-]?(\d*);[ \t]*download=[+-]?(\d*);[ \t]*total=[+-]?(\d*);[ \t]*expire=[+-]?(\d*)`)

func parseRawInfo(rawInfo string) (map[string]int64, bool) {
	result := rawInfoRegexp.FindStringSubmatch(rawInfo)

	if len(result) > 0 {
		upload, _ := strconv.ParseInt(result[1], 10, 64)
		download, _ := strconv.ParseInt(result[2], 10, 64)
		total, _ := strconv.ParseInt(result[3], 10, 64)
		expire, _ := strconv.ParseInt(result[4], 10, 64)

		info := map[string]int64{
			"Upload":   upload,
			"Download": download,
			"Total":    total,
			"Expire":   expire,
		}

		return info, true
	}
	return nil, false
}

func (p *MyProviderAdapter) SubInfo() map[string]int64 {
	p.access.RLock()
	defer p.access.RUnlock()
	if p.subInfo == nil {
		return nil
	}
	copyMap := make(map[string]int64, len(p.subInfo))
	for k, v := range p.subInfo {
		copyMap[k] = v
	}
	return copyMap
}

func (p *MyProviderAdapter) LastUpdateTime() time.Time {
	return p.lastUpdateTime
}

func (p *MyProviderAdapter) Outbound() adapter.OutboundManager {
	return p.outbound
}

func (p *MyProviderAdapter) Endpoint() adapter.EndpointManager {
	return p.endpoint
}

func (p *MyProviderAdapter) Provider() adapter.ProviderManager {
	return p.provider
}
