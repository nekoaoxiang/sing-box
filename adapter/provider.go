package adapter

import (
	"context"
	"time"

	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
)

type Provider interface {
	Lifecycle
	Tag() string
	Type() string

	LastUpdateTime() time.Time

	SubInfo() map[string]int64

	UpdateProvider() error

	Outbound() OutboundManager
	Endpoint() EndpointManager
	Provider() ProviderManager
}

type ProviderRegistry interface {
	option.ProviderOptionsRegistry
	CreateProvider(ctx context.Context, router Router, logger log.Factory, tag string, providerType string, options any) (Provider, error)
}

type ProviderManager interface {
	Lifecycle
	Providers() []Provider
	Provider(tag string) (Provider, bool)
	Remove(tag string) error
	Create(ctx context.Context, router Router, logger log.Factory, tag string, providerType string, options any) error
}
