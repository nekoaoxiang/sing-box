package dns

import (
	"context"
	"net"
	"net/netip"
	"net/url"
	"os"
	"sync"

	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/task"

	"github.com/miekg/dns"
)

var _ Upstream = (*UDPUpstream)(nil)

func init() {
	RegisterUpstream([]string{"udp", ""}, func(options UpstreamOptions) (Upstream, error) {
		return NewUDPUpstream(options)
	})
}

type UDPUpstream struct {
	name         string
	optCtx       context.Context
	ctx          context.Context
	cancel       context.CancelFunc
	dialer       N.Dialer
	logger       logger.ContextLogger
	serverAddr   M.Socksaddr
	clientAddr   netip.Prefix
	udpSize      int
	tcpUpstream *TCPUpstream
	access       sync.Mutex
	conn         *dnsConnection
}

func NewUDPUpstream(options UpstreamOptions) (*UDPUpstream, error) {
	var serverAddr M.Socksaddr
	if serverURL, err := url.Parse(options.Address); err != nil || serverURL.Scheme == "" {
		serverAddr = M.ParseSocksaddr(options.Address)
	} else {
		serverAddr = M.ParseSocksaddr(serverURL.Host)
	}
	if !serverAddr.IsValid() {
		return nil, E.New("invalid server address")
	}
	if serverAddr.Port == 0 {
		serverAddr.Port = 53
	}
	ctx, cancel := context.WithCancel(options.Context)
	return &UDPUpstream{
		name:         options.Name,
		optCtx:       options.Context,
		ctx:          ctx,
		cancel:       cancel,
		dialer:       options.Dialer,
		logger:       options.Logger,
		serverAddr:   serverAddr,
		clientAddr:   options.ClientSubnet,
		udpSize:      512,
		tcpUpstream: newTCPUpstream(options, serverAddr),
	}, nil
}

func (t *UDPUpstream) Name() string {
	return t.name
}

func (t *UDPUpstream) Start() error {
	return nil
}

func (t *UDPUpstream) Reset() {
	t.cancel()
	t.ctx, t.cancel = context.WithCancel(t.optCtx)
}

func (t *UDPUpstream) Close() error {
	t.cancel()
	return nil
}

func (t *UDPUpstream) Raw() bool {
	return true
}

func (t *UDPUpstream) Lookup(ctx context.Context, domain string, strategy DomainStrategy) ([]netip.Addr, error) {
	return nil, os.ErrInvalid
}

func (t *UDPUpstream) Exchange(ctx context.Context, message *dns.Msg) (*dns.Msg, error) {
	response, err := t.exchange(ctx, message)
	if err != nil {
		return nil, err
	}
	if response.Truncated {
		t.logger.InfoContext(ctx, "response truncated, retrying with TCP")
		return t.tcpUpstream.Exchange(ctx, message)
	}
	return response, nil
}

func (t *UDPUpstream) exchange(ctx context.Context, message *dns.Msg) (*dns.Msg, error) {
	conn, err := t.open(ctx)
	if err != nil {
		return nil, err
	}
	if edns0Opt := message.IsEdns0(); edns0Opt != nil {
		if udpSize := int(edns0Opt.UDPSize()); udpSize > t.udpSize {
			t.udpSize = udpSize
		}
	}
	buffer := buf.NewSize(1 + message.Len())
	defer buffer.Release()
	exMessage := *message
	exMessage.Compress = true
	messageId := message.Id
	callback := &dnsCallback{
		done: make(chan struct{}),
	}
	conn.access.Lock()
	conn.queryId++
	exMessage.Id = conn.queryId
	conn.callbacks[exMessage.Id] = callback
	conn.access.Unlock()
	defer func() {
		conn.access.Lock()
		delete(conn.callbacks, messageId)
		conn.access.Unlock()
		callback.access.Lock()
		select {
		case <-callback.done:
		default:
			close(callback.done)
		}
		callback.access.Unlock()
	}()
	rawMessage, err := exMessage.PackBuffer(buffer.FreeBytes())
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(rawMessage)
	if err != nil {
		conn.Close()
		return nil, err
	}
	select {
	case <-callback.done:
		callback.message.Id = messageId
		return callback.message, nil
	case <-conn.ctx.Done():
		return nil, E.Errors(conn.err, conn.ctx.Err())
	case <-ctx.Done():
		conn.Close()
		return nil, ctx.Err()
	}
}

func (t *UDPUpstream) open(ctx context.Context) (*dnsConnection, error) {
	connection := t.conn
	if connection != nil && !common.Done(connection.ctx) {
		return connection, nil
	}
	t.access.Lock()
	defer t.access.Unlock()
	connection = t.conn
	if connection != nil && !common.Done(connection.ctx) {
		return connection, nil
	}
	conn, err := t.dialer.DialContext(ctx, "udp", t.serverAddr)
	if err != nil {
		return nil, err
	}
	connCtx, cancel := context.WithCancel(t.ctx)
	connection = &dnsConnection{
		Conn:      conn,
		ctx:       connCtx,
		cancel:    cancel,
		callbacks: make(map[uint16]*dnsCallback),
	}
	t.conn = connection
	go t.recvLoop(connection)
	return connection, nil
}

func (t *UDPUpstream) recvLoop(conn *dnsConnection) {
	var group task.Group
	group.Append0(func(ctx context.Context) error {
		for {
			buffer := buf.NewSize(t.udpSize)
			_, err := buffer.ReadOnceFrom(conn)
			if err != nil {
				buffer.Release()
				return err
			}
			var message dns.Msg
			err = message.Unpack(buffer.Bytes())
			buffer.Release()
			if err != nil {
				return err
			}
			conn.access.RLock()
			callback, loaded := conn.callbacks[message.Id]
			conn.access.RUnlock()
			if !loaded {
				continue
			}
			callback.access.Lock()
			select {
			case <-callback.done:
			default:
				callback.message = &message
				close(callback.done)
			}
			callback.access.Unlock()
		}
	})
	group.Cleanup(func() {
		conn.Close()
	})
	group.Run(conn.ctx)
}

type dnsConnection struct {
	net.Conn
	ctx       context.Context
	cancel    context.CancelFunc
	access    sync.RWMutex
	err       error
	queryId   uint16
	callbacks map[uint16]*dnsCallback
}

func (c *dnsConnection) Close() error {
	c.cancel()
	return c.Conn.Close()
}

type dnsCallback struct {
	access  sync.Mutex
	message *dns.Msg
	done    chan struct{}
}
