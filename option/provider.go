package option

import (
	"context"

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
	err := json.UnmarshalContext(ctx, content, (*_Provider)(h))
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
