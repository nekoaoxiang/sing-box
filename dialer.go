package dns

import (
	"context"
	"net"
	"time"

	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

type DefaultDialer struct {
	dialer        N.Dialer
	client        *Client
	fallbackDelay time.Duration
}

func NewDefaultDialer(dialer N.Dialer, client *Client, fallbackDelay time.Duration) N.Dialer {
	return &DefaultDialer{dialer, client, fallbackDelay}
}

func (d *DefaultDialer) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if destination.IsIP() {
		return d.dialer.DialContext(ctx, network, destination)
	}
	destination.Fqdn = d.client.GetExactDomainFromHosts(ctx, destination.Fqdn, false)
	if addresses := d.client.GetAddrsFromHosts(ctx, destination.Fqdn, DomainStrategyAsIS, false); len(addresses) > 0 {
		return N.DialParallel(ctx, d.dialer, network, destination, addresses, false, d.fallbackDelay)
	}
	return nil, E.New("Invalid address")
}

func (d *DefaultDialer) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if destination.IsIP() {
		return d.dialer.ListenPacket(ctx, destination)
	}
	destination.Fqdn = d.client.GetExactDomainFromHosts(ctx, destination.Fqdn, false)
	if addresses := d.client.GetAddrsFromHosts(ctx, destination.Fqdn, DomainStrategyAsIS, false); len(addresses) > 0 {
		conn, _, err := N.ListenSerial(ctx, d.dialer, destination, addresses)
		return conn, err
	}
	return nil, E.New("Invalid address")
}

type DialerWrapper struct {
	DefaultDialer
	transport Transport
	strategy  DomainStrategy
}

func NewDialerWrapper(dialer N.Dialer, client *Client, transport Transport, strategy DomainStrategy, fallbackDelay time.Duration) N.Dialer {
	defaultDialer := DefaultDialer{dialer, client, fallbackDelay}
	return &DialerWrapper{defaultDialer, transport, strategy}
}

func (d *DialerWrapper) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	if destination.IsIP() {
		return d.dialer.DialContext(ctx, network, destination)
	}
	destination.Fqdn = d.client.GetExactDomainFromHosts(ctx, destination.Fqdn, false)
	if addresses := d.client.GetAddrsFromHosts(ctx, destination.Fqdn, d.strategy, false); len(addresses) > 0 {
		return N.DialParallel(ctx, d.dialer, network, destination, addresses, d.strategy == DomainStrategyPreferIPv6, d.fallbackDelay)
	}
	addresses, err := d.client.Lookup(ctx, d.transport, destination.Fqdn, d.strategy)
	if err != nil {
		return nil, err
	}
	return N.DialParallel(ctx, d.dialer, network, destination, addresses, d.strategy == DomainStrategyPreferIPv6, d.fallbackDelay)
}

func (d *DialerWrapper) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if destination.IsIP() {
		return d.dialer.ListenPacket(ctx, destination)
	}
	destination.Fqdn = d.client.GetExactDomainFromHosts(ctx, destination.Fqdn, false)
	if addresses := d.client.GetAddrsFromHosts(ctx, destination.Fqdn, DomainStrategyAsIS, false); len(addresses) > 0 {
		conn, _, err := N.ListenSerial(ctx, d.dialer, destination, addresses)
		return conn, err
	}
	addresses, err := d.client.Lookup(ctx, d.transport, destination.Fqdn, d.strategy)
	if err != nil {
		return nil, err
	}
	conn, _, err := N.ListenSerial(ctx, d.dialer, destination, addresses)
	return conn, err
}

func (d *DialerWrapper) Upstream() any {
	return d.dialer
}
