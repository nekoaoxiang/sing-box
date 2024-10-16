package adapter

import (
	"context"
	"time"

	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
)

type Provider interface {
	Tag() string
	Type() string

	SubInfo() map[string]int64

	PostStart() error

	LastUpdateTime() time.Time

	Outbound(tag string) (Outbound, bool)
	Outbounds() []Outbound

	Healthcheck(ctx context.Context, url string) map[string]uint16
	UpdateProvider(ctx context.Context) error
}

type ProviderRegistry interface {
	option.ProviderOptionsRegistry
	CreateProvider(ctx context.Context, router Router, logger log.ContextLogger, tag string, providerType string, options any) (Provider, error)
}

type ProviderManager interface {
	Lifecycle
	Providers() []Provider
	Provider(tag string) (Provider, bool)
	Remove(tag string) error
	Create(ctx context.Context, router Router, logger log.ContextLogger, tag string, providerType string, options any) error
}
