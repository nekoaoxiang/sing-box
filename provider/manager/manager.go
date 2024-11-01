package manager

import (
	"context"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing/common"
)

type Manager struct {
	ctx    context.Context
	logger log.ContextLogger
	router adapter.Router

	tag     string
	factory log.Factory

	provider adapter.ProviderManager
	outbound adapter.OutboundManager
	endpoint adapter.EndpointManager
}

func NewManager(ctx context.Context, logger log.ContextLogger, router adapter.Router, tag string, factory log.Factory) Manager {
	return Manager{
		ctx:    ctx,
		logger: logger,
		router: router,
		tag:    tag,

		factory: factory,
	}
}

func (m *Manager) NewOptions(options *Options) {
	m.NewManager()
	if options.Providers != nil {
		err := m.newProviders(options.Providers)
		if err != nil {
			m.logger.ErrorContext(m.ctx, "create provider failed: ", err)
		}
	}
	if options.Endpoints != nil {
		err := m.newEndpoints(options.Endpoints)
		if err != nil {
			m.logger.ErrorContext(m.ctx, "create endpoint failed: ", err)
		}
	}
	if options.Outbounds != nil {
		err := m.newOutbounds(options.Outbounds)
		if err != nil {
			m.logger.ErrorContext(m.ctx, "create outbound failed: ", err)
		}
	}
}

func (m *Manager) NewManager() error {
	m.newProviderManager()
	m.newEndpointManager()
	m.newOutboundManager()
	return nil
}

func (m *Manager) start(service adapter.Lifecycle) error {
	err := service.Start(adapter.StartStateInitialize)
	if err != nil {
		return err
	}
	err = service.Start(adapter.StartStateStart)
	if err != nil {
		return err
	}
	err = service.Start(adapter.StartStatePostStart)
	if err != nil {
		return err
	}
	err = service.Start(adapter.StartStateStarted)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) Outbound() adapter.OutboundManager {
	return m.outbound
}

func (m *Manager) Endpoint() adapter.EndpointManager {
	return m.endpoint
}

func (m *Manager) Provider() adapter.ProviderManager {
	return m.provider
}

func (m *Manager) Close() {
	common.Close(
		m.outbound, m.endpoint, m.provider,
	)
}
