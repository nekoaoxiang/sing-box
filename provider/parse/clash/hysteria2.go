package provider

import (
	"fmt"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func newClashHysteria2(proxy map[string]any) (*option.Outbound, error) {
	outbound := &option.Outbound{
		Type: C.TypeHysteria2,
	}

	options := &option.Hysteria2OutboundOptions{}
	obfsOptions := &option.Hysteria2Obfs{}

	if name, exists := proxy["name"].(string); exists {
		outbound.Tag = name
	}
	if server, exists := proxy["server"].(string); exists {
		options.Server = server
	}
	if port, exists := proxy["port"]; exists {
		options.ServerPort = stringToUint16(fmt.Sprint(port))
	}
	if password, exists := proxy["password"].(string); exists {
		options.Password = password
	}
	if up, exists := proxy["up"].(int); exists {
		options.UpMbps = up
	}
	if down, exists := proxy["down"].(int); exists {
		options.DownMbps = down
	}

	if obfs, exists := proxy["obfs"].(string); exists && obfs == "salamander" {
		obfsOptions.Type = obfs
	}
	if obfsPassword, exists := proxy["obfs-password"].(string); exists {
		obfsOptions.Password = obfsPassword
	}

	if obfsOptions.Type != "" {
		options.Obfs = obfsOptions
	}

	options.TLS = newTLSOptions(proxy)
	options.TLS.Enabled = true

	if ca, exists := proxy["ca"]; exists {
		options.TLS.CertificatePath = fmt.Sprint(ca)
	}
	if caStr, exists := proxy["ca-str"].([]any); exists {
		caStrArr := []string{}
		for _, item := range caStr {
			caStrArr = append(caStrArr, fmt.Sprint(item))
		}
		options.TLS.Certificate = caStrArr
	}

	options.DialerOptions = newDialerOptions(proxy)

	outbound.Options = *options
	return outbound, nil
}
