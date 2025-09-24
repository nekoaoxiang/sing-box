package provider

import (
	"strconv"

	E "github.com/sagernet/sing/common/exceptions"
	N "github.com/sagernet/sing/common/network"
	"gopkg.in/yaml.v3"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

var globalRegistry = &Registry{
	optionsType:  make(map[string]optionsConstructorFunc),
	constructors: make(map[string]constructorFunc),
}

func init() {
	Register[ShadowSocksOption](globalRegistry, "ss", newClashShadowsocks)
	Register[VMessOption](globalRegistry, C.TypeVMess, newClashVMess)
	Register[VlessOption](globalRegistry, C.TypeVLESS, newClashVLESS)
	Register[TrojanOption](globalRegistry, C.TypeTrojan, newClashTrojan)
	Register[AnyTLSOption](globalRegistry, C.TypeAnyTLS, newClashAnyTLS)
	Register[Hysteria2Option](globalRegistry, C.TypeHysteria2, newClashHysteria2)
}

func NewClashParser(content []byte) (*option.Options, error) {
	var config *Clash
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, E.Cause(err, "failed to unmarshal clash config: %w")
	}
	if len(config.Proxies) == 0 {
		return nil, E.New("no outbounds found in clash config")
	}

	outbounds := make([]option.Outbound, 0, len(config.Proxies))
	for _, proxy := range config.Proxies {
		outbound, err := globalRegistry.CreateOutbound(proxy.Name, proxy.Type, proxy.Options)
		if err != nil {
			return nil, err
		}
		outbounds = append(outbounds, *outbound)
	}

	options := &option.Options{
		Outbounds: outbounds,
	}
	return options, nil
}

func clashNetworks(udpEnabled *bool) option.NetworkList {
	if udpEnabled == nil || *udpEnabled {
		return ""
	} else {
		return N.NetworkTCP
	}
}

func stringToUint32(s string) uint32 {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(v)
}
