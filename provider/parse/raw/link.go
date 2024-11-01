package provider

import (
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/json/badoption"
)

var (
	linkPattern   = regexp.MustCompile(`^(?P<scheme>[^:]+)://(?P<payload>[^@?#]+)(?P<params>.*)$`)
	headerPattern = regexp.MustCompile(`^\s*(?P<key>[^:\s]+)\s*:\s*(?P<val>\S.*)$`)
	base64Pattern = regexp.MustCompile(`^[A-Za-z0-9+/=]+$`)
	wsPathPattern = regexp.MustCompile(`^(?P<path>[^?]+)(?:\?ed=(?P<ed>\d+))?$`)
)

// ParseSubscriptionLink parses a single subscription link into an Outbound config.
func ParseSubscriptionLink(linkUrl string) (*option.Outbound, error) {
	match := linkPattern.FindStringSubmatch(linkUrl)
	if match == nil {
		return nil, E.New("invalid subscription link format")
	}
	scheme := match[1]
	payload := match[2]
	params := match[3]

	// Decode Base64 payload if applicable
	if base64Pattern.MatchString(payload) && len(payload)%4 == 0 {
		if decoded, err := tryDecodeURLSafeBase64(payload); err == nil {
			payload = decoded
		}
	}

	// Reconstruct URL for parsing
	rebuilt := scheme + "://" + payload + params
	uri, err := url.Parse(rebuilt)
	if err != nil {
		return nil, E.Cause(err, "failed to parse URL")
	}

	switch scheme {
	case "ss":
		return newRawShadowsocks(uri)
	case C.TypeTUIC:
		return parseTuicLink(uri)
	case C.TypeVLESS:
		return parseVLESSLink(uri)
	case C.TypeTrojan:
		return parseTrojanLink(uri)
	case C.TypeHysteria:
		return parseHysteriaLink(uri)
	case "hy2", C.TypeHysteria2:
		return parseHysteria2Link(uri)
	default:
		return nil, E.New("unsupported scheme: ", scheme)
	}
}

// StringToType converts a string to a generic type T (int, uint, float, bool, string,
// badoption.Duration or badoption.HTTPHeader).
func StringToType[T any](s string) T {
	var zero T
	var rv = reflect.ValueOf(&zero).Elem()

	switch any(zero).(type) {
	case badoption.Duration:
		d, err := time.ParseDuration(s)
		if err == nil {
			rv.Set(reflect.ValueOf(badoption.Duration(d)))
		} else {
			i, _ := strconv.ParseInt(s, 10, 64)
			rv.SetInt(i)
		}
		return zero
	case badoption.HTTPHeader:
		h := badoption.HTTPHeader{}
		for _, line := range strings.Split(s, "\n") {
			if m := headerPattern.FindStringSubmatch(line); m != nil {
				h[m[1]] = strings.Split(m[2], ",")
			}
		}
		rv.Set(reflect.ValueOf(h))
		return zero
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, _ := strconv.ParseInt(s, 10, 64)
		rv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, _ := strconv.ParseUint(s, 10, 64)
		rv.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, _ := strconv.ParseFloat(s, 64)
		rv.SetFloat(f)
	case reflect.Bool:
		b, _ := strconv.ParseBool(s)
		rv.SetBool(b)
	case reflect.String:
		rv.SetString(s)
	default:
		panic("StringToType: unsupported type")
	}
	return zero
}

// v2rayTransportWs builds WebSocket options for V2Ray transports.
func v2rayTransportWs(host, rawPath string) option.V2RayWebsocketOptions {
	o := option.V2RayWebsocketOptions{}
	if host != "" {
		o.Headers = StringToType[badoption.HTTPHeader](F.ToString("Host: ", host))
	}
	if rawPath != "" {
		if m := wsPathPattern.FindStringSubmatch(rawPath); m != nil {
			o.Path = m[1]
			if m[2] != "" {
				o.EarlyDataHeaderName = "Sec-WebSocket-Protocol"
				o.MaxEarlyData = StringToType[uint32](m[2])
			}
		}
	}
	return o
}
