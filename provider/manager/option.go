package manager

import (
	"context"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json"
)

type Options struct {
	Type string `json:"type"`
	option.Options
}

// 兼容sing-box旧的outbound字段，移除待定
func (p *Options) UnmarshalJSONContext(ctx context.Context, content []byte) error {
	var raw struct {
		Outbounds []json.RawMessage
		Providers []option.Provider
		endpoints []option.Endpoint
	}
	err := json.UnmarshalContext(ctx, content, &raw)
	if err != nil {
		return err
	}
	for _, rawOutbound := range raw.Outbounds {
		var outbound option.Outbound

		var oldOutbound struct {
			Type string `json:"type"`
		}
		err := json.Unmarshal(rawOutbound, &oldOutbound)
		if err != nil {
			return err
		}
		switch oldOutbound.Type {
		case C.TypeBlock, C.TypeDNS:
			continue
		}

		err = outbound.UnmarshalJSONContext(ctx, rawOutbound)
		if err != nil {
			return err
		}
		p.Outbounds = append(p.Outbounds, outbound)
	}

	p.Providers = raw.Providers
	p.Endpoints = raw.endpoints
	return nil
}
