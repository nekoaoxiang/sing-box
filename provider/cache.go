package provider

import (
	"os"
	"time"

	"encoding/json"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

func (p *myProviderAdapter) parseCacheFile() (*option.SubInfo, time.Time, error) {
	fileInfo, err := os.Stat(p.path)
	if err != nil {
		return nil, time.Time{}, err
	}
	fileModeTime := fileInfo.ModTime()

	content, err := os.ReadFile(p.path)
	if err != nil {
		return nil, time.Time{}, err
	}

	RawOutbounds, info, err := p.parseCacheContent(content)
	if err != nil {
		return nil, time.Time{}, err
	}

	err = p.CreateOutbounds(p.ctx, p.router, RawOutbounds)
	if err != nil {
		return nil, time.Time{}, err
	}

	return info, fileModeTime, nil
}

func (p *myProviderAdapter) saveCacheContent(rawInfo *option.SubInfo, outbounds []option.Outbound) {
	cacheContentMap := make(map[string]interface{})

	if rawInfo != nil {
		cacheContentMap["info"] = rawInfo
	}

	var serializedOutbounds []json.RawMessage
	for _, outbound := range outbounds {
		switch outbound.Type {
		case C.TypeDirect, C.TypeBlock, C.TypeDNS, C.TypeSelector, C.TypeURLTest:
			continue
		}
		data, err := outbound.MarshalJSONContext(p.ctx)
		if err != nil {
			p.logger.ErrorContext(p.ctx, E.New("failed to marshal outbound to JSON: ", err))
			continue
		}
		serializedOutbounds = append(serializedOutbounds, data)
	}

	cacheContentMap["outbounds"] = serializedOutbounds

	finalContent, err := json.MarshalIndent(cacheContentMap, "", "  ")
	if err != nil {
		p.logger.ErrorContext(p.ctx, E.New("failed to marshal final content to JSON: ", err))
	}

	err = os.WriteFile(p.path, []byte(finalContent), 0o666)
	if err != nil {
		p.logger.ErrorContext(p.ctx, E.New("save provider", "[", p.Adapter.Tag(), "]", " failed: ", err))
	}
}

func (p *myProviderAdapter) parseCacheContent(content []byte) ([]option.Outbound, *option.SubInfo, error) {
	var options option.ProviderCacheOptions

	err := options.UnmarshalJSONContext(p.ctx, content)
	if err != nil {
		return nil, nil, E.Cause(err, "decode config at ")
	}
	outbounds := options.Outbounds
	info := options.Info
	return outbounds, info, nil
}
