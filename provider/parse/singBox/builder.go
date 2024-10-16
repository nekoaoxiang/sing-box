package provider

import (
	"context"

	"github.com/sagernet/sing-box/option"
	E "github.com/sagernet/sing/common/exceptions"
)

func NewSingBoxParser(ctx context.Context, content []byte) ([]option.Outbound, error) {
	var options option.ProviderCacheOptions
	err := options.UnmarshalJSONContext(ctx, content)
	if err != nil {
		return nil, E.Cause(err, "decode config at ")
	}
	outbounds := options.Outbounds
	return outbounds, nil
}
