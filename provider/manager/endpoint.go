package manager

import (
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/endpoint"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/service"
)

func (e *Manager) newEndpointManager() {
	endpointRegistry := service.FromContext[adapter.EndpointRegistry](e.ctx)
	e.endpoint = endpoint.NewManager(e.logger, endpointRegistry)
}

func (e *Manager) newEndpoints(options []option.Endpoint) error {
	e.logger.DebugContext(e.ctx, "create endpoints")
	err := e.createEndpoints(options)
	if err != nil {
		return err
	}

	return nil
}

func (p *Manager) createEndpoints(options []option.Endpoint) error {
	for i, endpointOptions := range options {
		var tag string
		if endpointOptions.Tag != "" {
			tag = endpointOptions.Tag
		} else {
			tag = F.ToString(i)
		}
		err := p.endpoint.Create(
			p.ctx,
			p.router,
			p.factory.NewLogger(F.ToString("provider[", p.tag, "]endpoint/[", tag, "]")),
			tag,
			endpointOptions.Type,
			endpointOptions.Options,
		)
		if err != nil {
			return E.Cause(err, "initialize endpoint[", i, "]")
		}
	}
	return nil
}

// func (p *Manager) startEndpoints() error {
// 	err := adapter.Start(adapter.StartStateStart, p.endpoint)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
