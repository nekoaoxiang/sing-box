package dns

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"

	"github.com/sagernet/sing/common/buf"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"

	"github.com/miekg/dns"
)

const MimeType = "application/dns-message"

var _ Upstream = (*HTTPSUpstream)(nil)

type HTTPSUpstream struct {
	destination string
	transport   *http.Transport
}

func init() {
	RegisterUpstream([]string{"https"}, func(options UpstreamOptions) (Upstream, error) {
		return NewHTTPSUpstream(options), nil
	})
}

func NewHTTPSUpstream(options UpstreamOptions) *HTTPSUpstream {
	return &HTTPSUpstream{
		destination: options.Address,
		transport: &http.Transport{
			ForceAttemptHTTP2: true,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return options.Dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
			},
			TLSClientConfig: &tls.Config{
				NextProtos: []string{"dns"},
			},
		},
	}
}

func (t *HTTPSUpstream) Start() error {
	return nil
}

func (t *HTTPSUpstream) Reset() {
	t.transport.CloseIdleConnections()
	t.transport = t.transport.Clone()
}

func (t *HTTPSUpstream) Close() error {
	t.Reset()
	return nil
}

func (t *HTTPSUpstream) Exchange(ctx context.Context, message *dns.Msg) (*dns.Msg, error) {
	exMessage := *message
	exMessage.Id = 0
	exMessage.Compress = true
	requestBuffer := buf.NewSize(1 + message.Len())
	rawMessage, err := exMessage.PackBuffer(requestBuffer.FreeBytes())
	if err != nil {
		requestBuffer.Release()
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, t.destination, bytes.NewReader(rawMessage))
	if err != nil {
		requestBuffer.Release()
		return nil, err
	}
	request.Header.Set("Content-Type", MimeType)
	request.Header.Set("Accept", MimeType)
	response, err := t.transport.RoundTrip(request)
	requestBuffer.Release()
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, E.New("unexpected status: ", response.Status)
	}
	var responseMessage dns.Msg
	if response.ContentLength > 0 {
		responseBuffer := buf.NewSize(int(response.ContentLength))
		_, err = responseBuffer.ReadFullFrom(response.Body, int(response.ContentLength))
		if err != nil {
			return nil, err
		}
		err = responseMessage.Unpack(responseBuffer.Bytes())
		responseBuffer.Release()
	} else {
		rawMessage, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		err = responseMessage.Unpack(rawMessage)
	}
	if err != nil {
		return nil, err
	}
	return &responseMessage, nil
}

func (t *HTTPSUpstream) Lookup(ctx context.Context, domain string, strategy DomainStrategy) ([]netip.Addr, error) {
	return nil, os.ErrInvalid
}
