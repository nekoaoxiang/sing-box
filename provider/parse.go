package provider

import (
	"encoding/json"

	"gopkg.in/yaml.v3"

	clash "github.com/sagernet/sing-box/provider/parse/clash"
	singBox "github.com/sagernet/sing-box/provider/parse/singBox"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

func detectDataFormat(data []byte) string {
	var jsonCheck interface{}
	if json.Unmarshal(data, &jsonCheck) == nil {
		return C.TypeSingBoxConfig
	}
	var yamlCheck interface{}
	if yaml.Unmarshal(data, &yamlCheck) == nil {
		return C.TpyeClashConfig
	}
	return ""
}

func (p *myProviderAdapter) newParser(content []byte) ([]option.Outbound, error) {
	var outbounds []option.Outbound
	var err error
	switch detectDataFormat(content) {
	case C.TypeSingBoxConfig:
		outbounds, err = singBox.NewSingBoxParser(p.ctx, content)
		if err != nil {
			return nil, err
		}
	case C.TpyeClashConfig:
		outbounds, err = clash.NewClashParser(content)
		if err != nil {
			return nil, err
		}
	default:
		return nil, E.Cause(err, "未知的配置: ")
	}
	return outbounds, nil
}
