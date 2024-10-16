package provider

import (
	"context"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

func NewSingBoxParser(ctx context.Context, content []byte) ([]option.Outbound, error) {
	outbounds := []option.Outbound{}
	var options option.ProviderCacheOptions
	err := options.UnmarshalJSONContext(ctx, content)
	if err != nil {
		return nil, E.Cause(err, "decode config at ")
	}
	for _, outbound := range options.Outbounds {
		switch outbound.Type {
		case C.TypeDirect, C.TypeBlock, C.TypeDNS, C.TypeSelector, C.TypeURLTest:
			continue
		default:
			outbounds = append(outbounds, outbound)
		}
	}
	return outbounds, nil
}
