package dns

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"os"
	"sync"

	"github.com/miekg/dns"
	"github.com/sagernet/sing/common/logger"
	N "github.com/sagernet/sing/common/network"
)

var _ Transport = (*BaseTransport)(nil)

type BaseTransport struct {
	name      string
	upstreams []Upstream
}

func (t *BaseTransport) Name() string {
	return t.name
}

func (t *BaseTransport) Start() error {
	for _, upstream := range t.upstreams {
		if err := upstream.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (t *BaseTransport) Reset() {
	for _, upstream := range t.upstreams {
		upstream.Reset()
	}
}

func (t *BaseTransport) Close() error {
	for _, upstream := range t.upstreams {
		if err := upstream.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (t *BaseTransport) Raw() bool {
	return true
}

type exchangeResult struct {
	msg *dns.Msg
	err error
}

func needBreak(message *dns.Msg, res *exchangeResult) bool {
	if errors.Is(res.err, context.DeadlineExceeded) {
		return true
	}
	if res.err != nil || res.msg.Rcode != dns.RcodeSuccess {
		return false
	}
	if message.Question[0].Qtype != dns.TypeA && message.Question[0].Qtype != dns.TypeAAAA {
		return true
	}
	var addrs []net.IP
	for _, result := range res.msg.Answer {
		switch record := result.(type) {
		case *dns.A:
			addrs = append(addrs, record.A)
		case *dns.AAAA:
			addrs = append(addrs, record.AAAA)
		}
	}
	return len(addrs) > 0
}

func (t *BaseTransport) Exchange(ctx context.Context, message *dns.Msg) (*dns.Msg, error) {
	var wg sync.WaitGroup
	results := make(chan exchangeResult, len(t.upstreams))
	for _, upstream := range t.upstreams {
		wg.Add(1)
		go func(upstream Upstream, ctx context.Context, message *dns.Msg, results chan exchangeResult, wg *sync.WaitGroup) {
			msg, err := upstream.Exchange(ctx, message)
			results <- exchangeResult{
				msg: msg,
				err: err,
			}
			wg.Done()
		}(upstream, ctx, message, results, &wg)
	}
	defer func(results chan exchangeResult, wg *sync.WaitGroup) {
		go func(results chan exchangeResult, wg *sync.WaitGroup) {
			wg.Wait()
			close(results)
		}(results, wg)
	}(results, &wg)
	var result *exchangeResult
	for i := 0; i < len(t.upstreams); i++ {
		res := <-results
		if !errors.Is(res.err, context.DeadlineExceeded) || result == nil {
			result = &res
		}
		if needBreak(message, &res) {
			break
		}
	}
	return result.msg, result.err
}

func (t *BaseTransport) Lookup(ctx context.Context, domain string, strategy DomainStrategy) ([]netip.Addr, error) {
	return nil, os.ErrInvalid
}

type Upstream interface {
	Start() error
	Reset()
	Close() error
	Exchange(ctx context.Context, message *dns.Msg) (*dns.Msg, error)
	Lookup(ctx context.Context, domain string, strategy DomainStrategy) ([]netip.Addr, error)
}

type UpstreamOptions struct {
	Context      context.Context
	Logger       logger.ContextLogger
	Name         string
	Dialer       N.Dialer
	Address      string
	ClientSubnet netip.Prefix
}

func baseTransportConstructor(options TransportOptions) (Transport, error) {
	var ups []Upstream
	for _, address := range options.Address {
		serverURL, _ := url.Parse(address)
		var scheme string
		if serverURL != nil {
			scheme = serverURL.Scheme
		}
		constructor := upstreams[scheme]
		upstream, err := constructor(UpstreamOptions{
			Context: options.Context,
			Logger:  options.Logger,
			Dialer:  options.Dialer,
			Address: address,
		})
		if err != nil {
			return nil, err
		}
		ups = append(ups, upstream)
	}
	return &BaseTransport{
		name:      options.Name,
		upstreams: ups,
	}, nil
}
