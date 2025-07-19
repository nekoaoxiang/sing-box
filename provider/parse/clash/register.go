package provider

import (
	"sync"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
)

type ConstructorFunc[T any] func(tag string, options T) (*option.Outbound, error)

func Register[Options any](registry *Registry, outboundType string, constructor ConstructorFunc[Options]) {
	registry.register(outboundType, func() any {
		return new(Options)
	}, func(tag string, rawOptions any) (*option.Outbound, error) {
		var options *Options
		if rawOptions != nil {
			options = rawOptions.(*Options)
		}
		return constructor(tag, common.PtrValueOrDefault(options))
	})
}

func (r *Registry) CreateOptions(outboundType string) (any, bool) {
	r.access.Lock()
	defer r.access.Unlock()
	optionsConstructor, loaded := r.optionsType[outboundType]
	if !loaded {
		return nil, false
	}
	return optionsConstructor(), true
}

func (r *Registry) CreateOutbound(tag, outboundType string, options any) (*option.Outbound, error) {
	r.access.Lock()
	defer r.access.Unlock()
	constructor, loaded := r.constructors[outboundType]
	if !loaded {
		return nil, E.New("outbound type not found: " + outboundType)
	}
	return constructor(tag, options)
}

type (
	optionsConstructorFunc func() any
	constructorFunc        func(tag string, options any) (*option.Outbound, error)
)

type Registry struct {
	access       sync.Mutex
	optionsType  map[string]optionsConstructorFunc
	constructors map[string]constructorFunc
}

func (r *Registry) register(outboundType string, optionsConstructor optionsConstructorFunc, constructor constructorFunc) {
	r.access.Lock()
	defer r.access.Unlock()
	r.optionsType[outboundType] = optionsConstructor
	r.constructors[outboundType] = constructor
}
