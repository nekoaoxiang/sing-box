package provider

import (
	"context"

	"github.com/sagernet/sing-box/provider/manager"
	clash "github.com/sagernet/sing-box/provider/parse/clash"
	raw "github.com/sagernet/sing-box/provider/parse/raw"
	singBox "github.com/sagernet/sing-box/provider/parse/singBox"

	E "github.com/sagernet/sing/common/exceptions"
)

func NewParser(ctx context.Context, content []byte) (*manager.Options, error) {
	var (
		options *manager.Options
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
