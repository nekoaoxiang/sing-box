package provider

import (
	"context"
	"os"
	"path/filepath"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
)

type _subInfo struct {
	Upload   int64 `json:"upload,omitempty"`
	Download int64 `json:"download,omitempty"`
	Total    int64 `json:"total,omitempty"`
	Expire   int64 `json:"expire,omitempty"`
}

type subInfo _subInfo

func (s *subInfo) UnmarshalJSONContext(ctx context.Context, content []byte) error {
	err := json.UnmarshalContext(ctx, content, (*_subInfo)(s))
	if err != nil {
		return err
	}
	return nil
}

type cacheOptions struct {
	Info subInfo `json:"info"`
}

func (r *Remote) parseCacheFile() error {
	fileInfo, err := os.Stat(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			// 没有缓存文件，正常返回（首次运行）
			r.logger.InfoContext(r.ctx, "cache file not found: ", r.path)
			return nil
		}
		// 其他 IO 错误，返回以便上层处理
		return err
	}
	r.lastUpdateTime = fileInfo.ModTime()
	content, err := os.ReadFile(r.path)
	if err != nil {
		r.logger.ErrorContext(r.ctx, "read cache file failed: ", err)
		return err
	}
	rawInfo, options, err := r.parseCacheContent(content)
	if err != nil {
		r.logger.ErrorContext(r.ctx, "parse config failed: ", err)
		return nil
	}
	if rawInfo != nil {
		info := make(map[string]int64)
		info["Upload"] = rawInfo.Upload
		info["Download"] = rawInfo.Download
		info["Total"] = rawInfo.Total
		info["Expire"] = rawInfo.Expire
		r.subInfo = info
	}
	r.UpdateOutbounds([]option.Outbound{}, options)

	r.logger.InfoContext(r.ctx, "loaded cache and applied outbounds, count=", len(options))
	return nil
}

func (p *Remote) saveCacheContent(rawInfo map[string]int64, options *option.Options) {
	cacheContentMap := make(map[string]any)
	if rawInfo != nil {
		cacheContentMap["info"] = rawInfo
	}

	if options.Outbounds != nil {
		var outbounds []json.RawMessage
		for _, outbound := range options.Outbounds {
			switch outbound.Type {
			case C.TypeDirect, C.TypeBlock, C.TypeDNS, C.TypeSelector, C.TypeURLTest:
				continue
			}
			data, err := outbound.MarshalJSONContext(p.ctx)
			if err != nil {
				p.logger.ErrorContext(p.ctx, E.New("failed to marshal outbound/", outbound.Type, "[", outbound.Tag, "] to JSON: ", err))
				continue
			}
			outbounds = append(outbounds, data)
		}
		cacheContentMap["outbounds"] = outbounds
	}

	if options.Providers != nil {
		cacheContentMap["providers"] = options.Providers
	}

	finalContent, err := json.MarshalContext(p.ctx, cacheContentMap)
	if err != nil {
		p.logger.ErrorContext(p.ctx, E.New("failed to marshal final content to JSON: ", err))
	}

	err = os.MkdirAll(filepath.Dir(p.path), 0o755)
	if err != nil {
		p.logger.ErrorContext(p.ctx, E.New("mkdir failed: ", err))
	}
	err = os.WriteFile(p.path, finalContent, 0o666)
	if err != nil {
		p.logger.ErrorContext(p.ctx, E.New("save file failed: ", err))
	}
}

func (p *Remote) parseCacheContent(content []byte) (*subInfo, []option.Outbound, error) {
	options, err := json.UnmarshalExtendedContext[cacheOptions](p.ctx, content)
	if err != nil {
		return nil, nil, E.Cause(err, "decode config at ", p.path)
	}

	var raw struct {
		Outbounds []json.RawMessage
	}
	err = json.UnmarshalContext(p.ctx, content, &raw)
	if err != nil {
		return nil, nil, err
	}
	var outbounds []option.Outbound
	for i, raw := range raw.Outbounds {
		var ob option.Outbound
		if err := ob.UnmarshalJSONContext(p.ctx, raw); err != nil {
			p.logger.WarnContext(
				p.ctx,
				"failed to unmarshal outbound #", i, ": ", err,
			)
			continue
		}
		outbounds = append(outbounds, ob)
	}

	return &options.Info, outbounds, nil
}
