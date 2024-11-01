package provider

import (
	"fmt"
	"strconv"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/provider/manager"

	"gopkg.in/yaml.v3"
)

type ClashConfig struct {
	Proxies []map[string]any `yaml:"proxies"`
}

func NewClashParser(raw []byte) (*manager.Options, error) {
	var config ClashConfig
	err := yaml.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}
	if len(config.Proxies) == 0 {
		return nil, fmt.Errorf("no outbounds found in clash config")
	}

	var outbounds []option.Outbound
	for _, proxy := range config.Proxies {
		var (
			outbound *option.Outbound
			err      error
		)

		protocol, exists := proxy["type"]
		if !exists {
			continue
		}

		switch protocol {
		case "ss":
			outbound, err = newClashShadowsocks(proxy)
		case C.TypeVMess:
			outbound, err = newClashVMess(proxy)
		case C.TypeTrojan:
			outbound, err = newClashTrojan(proxy)
		case C.TypeVLESS:
			outbound, err = newClashVLESS(proxy)
		case C.TypeHysteria2:
			outbound, err = newClashHysteria2(proxy)
		default:
			continue
		}
		if err == nil {
			outbounds = append(outbounds, *outbound)
		}
	}

	options := &manager.Options{
		Type: C.TpyeClashConfig,
		Options: option.Options{
			Outbounds: outbounds,
		},
	}
	return options, nil
}

func stringToUint16(content string) uint16 {
	intNum, _ := strconv.Atoi(content)
	return uint16(intNum)
}

func stringToUint32(content string) uint32 {
	intNum, _ := strconv.Atoi(content)
	return uint32(intNum)
}
