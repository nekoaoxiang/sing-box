package option

import (
	"context"

	C "github.com/sagernet/sing-box/constant"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/json"
	"github.com/sagernet/sing/common/json/badjson"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/sagernet/sing/service"
)

type ProviderOptionsRegistry interface {
	CreateOptions(providerType string) (any, bool)
}

type _Provider struct {
	Type    string `json:"type"`
	Tag     string `json:"tag,omitempty"`
	Options any    `json:"-"`
}

type Provider _Provider

func (h *Provider) MarshalJSONContext(ctx context.Context) ([]byte, error) {
	return badjson.MarshallObjectsContext(ctx, (*_Provider)(h), h.Options)
}

func (h *Provider) UnmarshalJSONContext(ctx context.Context, content []byte) error {
	err := json.Unmarshal(content, (*_Provider)(h))
	if err != nil {
		return err
	}
	registry := service.FromContext[ProviderOptionsRegistry](ctx)
	if registry == nil {
		return E.New("missing provider options registry in context")
	}
	options, loaded := registry.CreateOptions(h.Type)
	if !loaded {
		return E.New("unknown provider type: ", h.Type)
	}
	err = badjson.UnmarshallExcludedContext(ctx, content, (*_Provider)(h), options)
	if err != nil {
		return err
	}
	h.Options = options
	return nil
}

type ProviderOptions struct {
	Type             string                   `json:"type"`
	Path             string                   `json:"path,omitempty"`
	Tag              string                   `json:"tag,omitempty"`
	OutboundOverride *OutboundOverrideOptions `json:"outbound_override,omitempty"`
	HealthCheck      *HealthCheckOptions      `json:"health_check,omitempty"`
	Filter           *FilterOptions           `json:"filter,omitempty"`
}

type RemoteProviderOptions struct {
	ProviderOptions
	Url       string             `json:"download_url"`
	UserAgent string             `json:"download_ua,omitempty"`
	Interval  badoption.Duration `json:"download_interval,omitempty"`
	Detour    string             `json:"download_detour,omitempty"`
}

type LocalProviderOptions struct {
	ProviderOptions
}

type HealthCheckOptions struct {
	Enable   bool               `json:"enable,omitempty"`
	Url      string             `json:"url,omitempty"`
	Interval badoption.Duration `json:"interval,omitempty"`
}

type OutboundOverrideOptions struct {
	TagPrefix string `json:"tag_prefix,omitempty"`
	TagSuffix string `json:"tag_suffix,omitempty"`
	*OverrideDialerOptions
}

type OverrideDialerOptions struct {
	Detour           *string             `json:"detour,omitempty"`
	BindInterface    *string             `json:"bind_interface,omitempty"`
	Inet4BindAddress *badoption.Addr     `json:"inet4_bind_address,omitempty"`
	Inet6BindAddress *badoption.Addr     `json:"inet6_bind_address,omitempty"`
	ProtectPath      *string             `json:"protect_path,omitempty"`
	RoutingMark      *uint32             `json:"routing_mark,omitempty"`
	ReuseAddr        *bool               `json:"reuse_addr,omitempty"`
	ConnectTimeout   *badoption.Duration `json:"connect_timeout,omitempty"`
	TCPFastOpen      *bool               `json:"tcp_fast_open,omitempty"`
	TCPMultiPath     *bool               `json:"tcp_multi_path,omitempty"`
	UDPFragment      *bool               `json:"udp_fragment,omitempty"`
	DomainStrategy   *DomainStrategy     `json:"domain_strategy,omitempty"`
	FallbackDelay    *badoption.Duration `json:"fallback_delay,omitempty"`
}

// func (h Provider) MarshalJSON() ([]byte, error) {
// 	var v any
// 	switch h.Type {
// 	case C.ProviderTypeLocal:
// 		v = h
// 	case C.ProviderTypeRemote:
// 		v = h.RemoteOptions
// 	default:
// 		return nil, E.New("unknown provider type: ", h.Type)
// 	}
// 	return badjson.MarshallObjects((_Provider)(h), v)
// }

// func (h *Provider) UnmarshalJSON(bytes []byte) error {
// 	err := json.Unmarshal(bytes, (*_Provider)(h))
// 	if err != nil {
// 		return err
// 	}
// 	var v any
// 	switch h.Type {
// 	case C.ProviderTypeLocal:
// 		v = &h
// 	case C.ProviderTypeRemote:
// 		v = &h.RemoteOptions
// 	default:
// 		return E.New("unknown provider type: ", h.Type)
// 	}
// 	err = badjson.UnmarshallExcluded(bytes, (*_Provider)(h), v)
// 	if err != nil {
// 		return E.Cause(err, "provider options")
// 	}
// 	return nil
// }

type SubInfo struct {
	Upload   int64 `json:"upload,omitempty"`
	Download int64 `json:"download,omitempty"`
	Total    int64 `json:"total,omitempty"`
	Expire   int64 `json:"expire,omitempty"`
}

type _ProviderCacheOptions struct {
	Info      *SubInfo   `json:"info,omitempty"`
	Outbounds []Outbound `json:"outbounds"`
}

type ProviderCacheOptions _ProviderCacheOptions

// func (p *ProviderCacheOptions) UnmarshalJSONContext(ctx context.Context, content []byte) error {
// 	err := json.Unmarshal(content, (*_ProviderCacheOptions)(p))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (p *ProviderCacheOptions) UnmarshalJSONContext(ctx context.Context, content []byte) error {
	var raw struct {
		Info      *SubInfo          `json:"info,omitempty"`
		Outbounds []json.RawMessage `json:"outbounds"`
	}

	// 先解码到一个临时结构体
	err := json.Unmarshal(content, &raw)
	if err != nil {
		return err
	}

	p.Info = raw.Info

	// 手动处理 Outbounds 字段
	for _, rawOutbound := range raw.Outbounds {
		var outbound Outbound

		// 兼容sing-box旧的outbound字段，移除待定
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

	return nil
}
