package provider

import (
	"fmt"
	"regexp"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

func newV2RayTransport(proxy map[string]any) *option.V2RayTransportOptions {
	if network, exists := proxy["network"].(string); exists {
		Transport := option.V2RayTransportOptions{}
		switch network {
		case "http":
			Transport.Type = C.V2RayTransportTypeHTTP
			Transport.HTTPOptions = newHTTPTransport(proxy)
		case "ws":
			Transport.Type = C.V2RayTransportTypeWebsocket
			Transport.WebsocketOptions = newWebsocketTransport(proxy)
		case "grpc":
			Transport.Type = C.V2RayTransportTypeGRPC
			Transport.GRPCOptions = newGRPCTransport(proxy)
		}
		return &Transport
	}
	return nil
}

func newHTTPTransport(proxy map[string]any) option.V2RayHTTPOptions {
	options := option.V2RayHTTPOptions{
		Host:    badoption.Listable[string]{},
		Headers: map[string]badoption.Listable[string]{},
	}
	if httpOpts, exists := proxy["http-opts"].(map[string]any); exists {
		if hostsRaw, exists := httpOpts["host"]; exists {
			switch hosts := hostsRaw.(type) {
			case []string:
				options.Host = hosts
			case string:
				options.Host = []string{hosts}
			}
		}
		if pathRaw, exists := httpOpts["path"]; exists {
			switch path := pathRaw.(type) {
			case []string:
				options.Path = path[0]
			case string:
				options.Path = path
			}
		}
		if method, exists := httpOpts["method"].(string); exists {
			options.Method = method
		}
		if headers, exists := httpOpts["headers"].(map[string]any); exists {
			for key, valueRaw := range headers {
				valueArr := []string{}
				switch value := valueRaw.(type) {
				case []any:
					for _, item := range value {
						valueArr = append(valueArr, fmt.Sprint(item))
					}
				default:
					valueArr = append(valueArr, fmt.Sprint(value))
				}
				options.Headers[key] = valueArr
			}
		}
	}
	return options
}

func newWebsocketTransport(proxy map[string]any) option.V2RayWebsocketOptions {
	options := option.V2RayWebsocketOptions{
		Headers: map[string]badoption.Listable[string]{},
	}
	if wsOpts, exists := proxy["ws-opts"].(map[string]any); exists {
		if path, exists := wsOpts["path"].(string); exists {
			reg := regexp.MustCompile(`^(.*?)(?:\?ed=(\d+))?$`)
			result := reg.FindStringSubmatch(path)
			options.Path = result[1]
			if result[2] != "" {
				options.MaxEarlyData = stringToUint32(result[2])
				options.EarlyDataHeaderName = "Sec-WebSocket-Protocol"
			}
		}
		if headers, exists := wsOpts["headers"].(map[string]any); exists {
			for key, valueRaw := range headers {
				valueArr := []string{}
				switch value := valueRaw.(type) {
				case []any:
					for _, item := range value {
						valueArr = append(valueArr, fmt.Sprint(item))
					}
				default:
					valueArr = append(valueArr, fmt.Sprint(value))
				}
				options.Headers[key] = valueArr
			}
		}
		if maxEarlyData, exists := wsOpts["max-early-data"].(int); exists {
			options.MaxEarlyData = uint32(maxEarlyData)
		}
		if earlyDataHeaderName, exists := wsOpts["early-data-header-name"].(string); exists {
			options.EarlyDataHeaderName = earlyDataHeaderName
		}
	}
	if path, exists := proxy["ws-path"].(string); exists {
		reg := regexp.MustCompile(`^(.*?)(?:\?ed=(\d+))?$`)
		result := reg.FindStringSubmatch(path)
		options.Path = result[1]
		if result[2] != "" {
			options.MaxEarlyData = stringToUint32(result[2])
			options.EarlyDataHeaderName = "Sec-WebSocket-Protocol"
		}
	}
	if headers, exists := proxy["ws-headers"].(map[string]any); exists {
		for key, valueRaw := range headers {
			valueArr := []string{}
			switch value := valueRaw.(type) {
			case []any:
				for _, item := range value {
					valueArr = append(valueArr, fmt.Sprint(item))
				}
			default:
				valueArr = append(valueArr, fmt.Sprint(value))
			}
			options.Headers[key] = valueArr
		}
	}
	return options
}

func newGRPCTransport(proxy map[string]any) option.V2RayGRPCOptions {
	options := option.V2RayGRPCOptions{}
	if grpcOpts, exists := proxy["grpc-opts"].(map[string]any); exists {
		if servername, exists := grpcOpts["grpc-service-name"].(string); exists {
			options.ServiceName = servername
		}
	}
	return options
}
