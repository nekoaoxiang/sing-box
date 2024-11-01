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
	"github.com/sagernet/sing-box/protocol/group"
	"github.com/sagernet/sing-box/provider/manager"
)

type MyProviderAdapter struct {
	provider.Adapter
	manager.Manager

	ctx    context.Context
	logger log.ContextLogger
	access sync.RWMutex

	path string

	includes *group.FilterList
	excludes *group.FilterList

	lastUpdateTime time.Time
	subInfo        map[string]int64
}

func parseRawInfo(rawInfo string) (map[string]int64, bool) {
	reg := regexp.MustCompile(`upload=[+-]?(\d*);[ \t]*download=[+-]?(\d*);[ \t]*total=[+-]?(\d*);[ \t]*expire=[+-]?(\d*)`)
	result := reg.FindStringSubmatch(rawInfo)

	if len(result) > 0 {
		upload, _ := strconv.ParseInt(result[1], 10, 64)
		download, _ := strconv.ParseInt(result[2], 10, 64)
		total, _ := strconv.ParseInt(result[3], 10, 64)
		expire, _ := strconv.ParseInt(result[4], 10, 64)

		info := make(map[string]int64)
		info["Upload"] = upload
		info["Download"] = download
		info["Total"] = total
		info["Expire"] = expire

		return info, true
	}
	return nil, false
}

func (p *MyProviderAdapter) SubInfo() map[string]int64 {
	return p.subInfo
}

func (p *MyProviderAdapter) LastUpdateTime() time.Time {
	return p.lastUpdateTime
}

func (p *MyProviderAdapter) Outbound() adapter.OutboundManager {
	return p.Manager.Outbound()
}

func (p *MyProviderAdapter) Endpoint() adapter.EndpointManager {
	return p.Manager.Endpoint()
}

func (p *MyProviderAdapter) Provider() adapter.ProviderManager {
	return p.Manager.Provider()
}

func (p *MyProviderAdapter) Close() {
	p.Manager.Close()
}
