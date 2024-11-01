package provider

import (
	"net/url"
	"strconv"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/byteformats"
	E "github.com/sagernet/sing/common/exceptions"
)

func parseHysteriaLink(linkURL *url.URL) (*option.Outbound, error) {
	var options option.HysteriaOutboundOptions
	TLSOptions := option.OutboundTLSOptions{
		Enabled: true,
		ECH:     &option.OutboundECHOptions{},
		UTLS:    &option.OutboundUTLSOptions{},
		Reality: &option.OutboundRealityOptions{},
	}
	options.Server = linkURL.Hostname()
	TLSOptions.ServerName = linkURL.Hostname()
	options.ServerPort = StringToType[uint16](linkURL.Port())
	for key, values := range linkURL.Query() {
		value := values[0]
		switch key {
		case "auth":
			options.AuthString = value
		case "peer", "sni":
			TLSOptions.ServerName = value
		case "alpn":
			TLSOptions.ALPN = strings.Split(value, ",")
		case "ca":
			TLSOptions.CertificatePath = value
		case "ca_str":
			TLSOptions.Certificate = strings.Split(value, "\n")
		case "up":
			var up byteformats.NetworkBytesCompat
			if err := up.UnmarshalJSON([]byte(value)); err != nil {
				return nil, E.Cause(err, "invalid up value")
			}
			options.Up = &up
		case "up_mbps":
			options.UpMbps, _ = strconv.Atoi(value)
		case "down":
			var down byteformats.NetworkBytesCompat
			if err := down.UnmarshalJSON([]byte(value)); err != nil {
				return nil, E.Cause(err, "invalid down value")
			}
			options.Down = &down
		case "down_mbps":
			options.DownMbps, _ = strconv.Atoi(value)
		case "obfs", "obfsParam":
			options.Obfs = value
		case "insecure", "skip-cert-verify":
			if value == "1" || value == "true" {
				TLSOptions.Insecure = true
			}
		case "tfo", "tcp-fast-open", "tcp_fast_open":
			if value == "1" || value == "true" {
				options.TCPFastOpen = true
			}
		}
	}
	outbound := &option.Outbound{
		Type: C.TypeHysteria,
		Tag:  linkURL.Fragment,
	}
	options.TLS = &TLSOptions
	outbound.Options = &options
	return outbound, nil
}
