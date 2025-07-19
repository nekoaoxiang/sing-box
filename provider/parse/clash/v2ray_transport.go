package provider

import (
	"regexp"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

type V2RayTransportOption struct {
	HTTPOpts  HTTPOptions  `yaml:"http-opts,omitempty"`
	HTTP2Opts HTTP2Options `yaml:"h2-opts,omitempty"`
	GrpcOpts  GrpcOptions  `yaml:"grpc-opts,omitempty"`
	WSOpts    WSOptions    `yaml:"ws-opts,omitempty"`
}

type HTTPOptions struct {
	Method  string              `yaml:"method,omitempty"`
	Path    []string            `yaml:"path,omitempty"`
	Headers map[string][]string `yaml:"headers,omitempty"`
}

type HTTP2Options struct {
	Host []string `yaml:"host,omitempty"`
	Path string   `yaml:"path,omitempty"`
}

type GrpcOptions struct {
	GrpcServiceName string `yaml:"grpc-service-name,omitempty"`
}

type WSOptions struct {
	Path                     string            `yaml:"path,omitempty"`
	Headers                  map[string]string `yaml:"headers,omitempty"`
	MaxEarlyData             int               `yaml:"max-early-data,omitempty"`
	EarlyDataHeaderName      string            `yaml:"early-data-header-name,omitempty"`
	V2rayHttpUpgrade         bool              `yaml:"v2ray-http-upgrade,omitempty"`
	V2rayHttpUpgradeFastOpen bool              `yaml:"v2ray-http-upgrade-fast-open,omitempty"`
}

func newV2RayTransport(network string, proxy *V2RayTransportOption) *option.V2RayTransportOptions {
	if proxy == nil {
		return nil
	}

	Transport := &option.V2RayTransportOptions{}
	switch network {
	case C.V2RayTransportTypeHTTP:
		Transport.Type = C.V2RayTransportTypeHTTP
		Transport.HTTPOptions = newHTTPTransport(proxy.HTTPOpts)

	case C.V2RayTransportTypeWebsocket:
		wsOpts, isHTTPUpgrade := newWebsocketTransport(proxy.WSOpts)
		if isHTTPUpgrade {
			Transport.Type = C.V2RayTransportTypeHTTPUpgrade
			Transport.HTTPUpgradeOptions = newHTTPUpgradeTransport(proxy.WSOpts)
		} else {
			Transport.Type = C.V2RayTransportTypeWebsocket
			Transport.WebsocketOptions = wsOpts
		}

	case C.V2RayTransportTypeGRPC:
		Transport.Type = C.V2RayTransportTypeGRPC
		Transport.GRPCOptions = newGRPCTransport(proxy.GrpcOpts)
	}
	return Transport
}

func newHTTPTransport(proxy HTTPOptions) option.V2RayHTTPOptions {
	options := option.V2RayHTTPOptions{
		Headers: copyHeaders(proxy.Headers),
	}
	options.Method = proxy.Method
	if len(proxy.Path) > 0 {
		options.Path = proxy.Path[0]
	}
	return options
}

func newWebsocketTransport(proxy WSOptions) (option.V2RayWebsocketOptions, bool) {
	options := option.V2RayWebsocketOptions{
		Headers: copyWSHeaders(proxy.Headers),
	}
	options.Path, options.MaxEarlyData = parseWSPath(proxy.Path)
	if options.MaxEarlyData != 0 {
		options.EarlyDataHeaderName = "Sec-WebSocket-Protocol"
	}
	return options, proxy.V2rayHttpUpgrade
}

func newHTTPUpgradeTransport(proxy WSOptions) option.V2RayHTTPUpgradeOptions {
	options := option.V2RayHTTPUpgradeOptions{
		Headers: copyWSHeaders(proxy.Headers),
	}
	options.Path, _ = parseWSPath(proxy.Path)
	return options
}

func newGRPCTransport(proxy GrpcOptions) option.V2RayGRPCOptions {
	options := option.V2RayGRPCOptions{
		ServiceName: proxy.GrpcServiceName,
	}
	return options
}

func copyHeaders(src map[string][]string) map[string]badoption.Listable[string] {
	dst := make(map[string]badoption.Listable[string], len(src))
	for k, v := range src {
		dst[k] = append(dst[k], v...)
	}
	return dst
}

func copyWSHeaders(src map[string]string) map[string]badoption.Listable[string] {
	dst := make(map[string]badoption.Listable[string], len(src))
	for k, v := range src {
		dst[k] = append(dst[k], v)
	}
	return dst
}

func parseWSPath(path string) (realPath string, maxEarlyData uint32) {
	reg := regexp.MustCompile(`^(.*?)(?:\?ed=(\d+))?$`)
	result := reg.FindStringSubmatch(path)
	if len(result) < 2 {
		return path, 0
	}
	realPath = result[1]
	if len(result) >= 3 && result[2] != "" {
		maxEarlyData = stringToUint32(result[2])
	}
	return
}
