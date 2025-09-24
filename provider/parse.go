package provider

import (
	"context"

	// "github.com/sagernet/sing-box/provider/manager"
	"github.com/sagernet/sing-box/option"
	clash "github.com/sagernet/sing-box/provider/parse/clash"
	raw "github.com/sagernet/sing-box/provider/parse/raw"
	singBox "github.com/sagernet/sing-box/provider/parse/singBox"

	E "github.com/sagernet/sing/common/exceptions"
)

func NewParser(ctx context.Context, content []byte) (*option.Options, error) {
	var (
		options *option.Options
		err     error
	)
	options, err = singBox.NewSingBoxParser(ctx, content)
	if err == nil {
		return options, nil
	}
	options, err = clash.NewClashParser(content)
	if err == nil {
		return options, nil
	}
	options, err = raw.NewRawParser(string(content))
	if err == nil {
		return options, nil
	}
	return nil, E.Cause(err, "decode config at")
}
