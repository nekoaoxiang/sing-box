package provider

import (
	"context"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/provider/manager"
	"github.com/sagernet/sing/common/json"
)

func NewSingBoxParser(ctx context.Context, content []byte) (*manager.Options, error) {
	options, err := json.UnmarshalExtendedContext[manager.Options](ctx, content)
	if err != nil {
		return nil, err
	}

	var outbounds []option.Outbound
	for _, outbound := range options.Outbounds {
		switch outbound.Type {
		case C.TypeDirect, C.TypeBlock, C.TypeDNS, C.TypeSelector, C.TypeURLTest:
		default:
			outbounds = append(outbounds, outbound)
		}
	}
	options.Outbounds = outbounds
	options.Type = C.TypeSingBoxConfig

	return &options, nil
}
