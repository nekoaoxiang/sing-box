package provider

import (
	"os"
	"time"

	"encoding/json"
	"regexp"
	"strconv"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

// 解析缓存文件
func (p *myProviderAdapter) parseCacheFile() ([]adapter.Outbound, *option.SubInfo, time.Time, error) {
	fileInfo, err := os.Stat(p.path)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	fileModeTime := fileInfo.ModTime()

	content, err := os.ReadFile(p.path)
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	RawOutbounds, info, err := p.parseCacheContent(content)
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	outbounds, err := p.CreateOutbounds(p.ctx, p.router, RawOutbounds)
	if err != nil {
		return nil, nil, time.Time{}, err
	}

	return outbounds, info, fileModeTime, nil
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

// 解析缓存文件中出站部分，返回 []adapter.Outbound 类型
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

// 解析provider信息
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
