package provider

import (
	"context"
	"os"
	"path/filepath"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/provider/manager"
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
	Info subInfo `json:"info,omitempty"`
}

func (r *Remote) parseCacheFile() error {
	fileInfo, err := os.Stat(r.path)
	if err != nil {
		return err
	}
	r.lastUpdateTime = fileInfo.ModTime()
	content, err := os.ReadFile(r.path)
	if err != nil {
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
	r.NewOptions(options)
	return nil
}

func (p *Remote) saveCacheContent(rawInfo map[string]int64, options *manager.Options) {
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

func (p *Remote) parseCacheContent(content []byte) (*subInfo, *manager.Options, error) {
	info, err := json.UnmarshalExtendedContext[cacheOptions](p.ctx, content)
	if err != nil {
		return nil, nil, E.Cause(err, "decode config at ", p.path)
	}

	options, err := json.UnmarshalExtendedContext[manager.Options](p.ctx, content)
	if err != nil {
		return nil, nil, E.Cause(err, "decode config at ", p.path)
	}
	return &info.Info, &options, nil
}
