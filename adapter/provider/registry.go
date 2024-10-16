package provider

import (
	"context"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"

	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
)

type ConstructorFunc[T any] func(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options T) (adapter.Provider, error)

func Register[Options any](registry *Registry, providerType string, constructor ConstructorFunc[Options]) {
	registry.register(providerType, func() any {
		return new(Options)
	}, func(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options any) (adapter.Provider, error) {
		return constructor(ctx, router, logger, tag, common.PtrValueOrDefault(options.(*Options)))
	})
}

var _ adapter.ProviderRegistry = (*Registry)(nil)

type (
	optionsConstructorFunc func() any
	constructorFunc        func(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options any) (adapter.Provider, error)
)

type Registry struct {
	access       sync.Mutex
	optionsType  map[string]optionsConstructorFunc
	constructors map[string]constructorFunc
}

func NewRegistry() *Registry {
	return &Registry{
		optionsType:  make(map[string]optionsConstructorFunc),
		constructors: make(map[string]constructorFunc),
	}
}

func (r *Registry) CreateOptions(providerType string) (any, bool) {
	r.access.Lock()
	defer r.access.Unlock()
	optionsConstructor, loaded := r.optionsType[providerType]
	if !loaded {
		return nil, false
	}
	return optionsConstructor(), true
}

func (r *Registry) CreateProvider(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, providerType string, options any) (adapter.Provider, error) {
	r.access.Lock()
	defer r.access.Unlock()
	constructor, loaded := r.constructors[providerType]
	if !loaded {
		return nil, E.New("provider type not found: " + providerType)
	}
	return constructor(ctx, router, logger, tag, options)
}

func (r *Registry) register(providerType string, optionsConstructor optionsConstructorFunc, constructor constructorFunc) {
	r.access.Lock()
	defer r.access.Unlock()
	r.optionsType[providerType] = optionsConstructor
	r.constructors[providerType] = constructor
}
