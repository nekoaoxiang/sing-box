package provider

import (
	"net/url"
	"strconv"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func parseHysteria2Link(link *url.URL) (*option.Outbound, error) {
	options := option.Hysteria2OutboundOptions{
		Obfs: &option.Hysteria2Obfs{},
	}
	TLSOptions := option.OutboundTLSOptions{
		Enabled: true,
		ECH:     &option.OutboundECHOptions{},
		UTLS:    &option.OutboundUTLSOptions{},
		Reality: &option.OutboundRealityOptions{},
	}
	options.ServerPort = uint16(443)
	options.Server = link.Hostname()
	TLSOptions.ServerName = link.Hostname()
	if link.User != nil {
		options.Password = link.User.Username()
	}
	if link.Port() != "" {
		options.ServerPort = StringToType[uint16](link.Port())
	}
	for key, values := range link.Query() {
		value := values[0]
		switch key {
		case "up":
			options.UpMbps, _ = strconv.Atoi(value)
		case "down":
			options.DownMbps, _ = strconv.Atoi(value)
		case "obfs":
			if value == "salamander" {
				options.Obfs.Type = "salamander"
			}
		case "obfs-password":
			options.Obfs.Password = value
		case "insecure", "skip-cert-verify":
			if value == "1" || value == "true" {
				TLSOptions.Insecure = true
			}
		}
	}
	outbound := &option.Outbound{
		Type: C.TypeHysteria2,
		Tag:  link.Fragment,
	}
	options.TLS = &TLSOptions
	outbound.Options = &options
	return outbound, nil
}
