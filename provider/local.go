package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sagernet/fswatch"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
)

func RegisterLocal(registry *provider.Registry) {
	provider.Register[option.LocalProviderOptions](registry, C.TypeLocal, NewLocal)
}

var (
	_ adapter.Provider = (*Local)(nil)
)

type Local struct {
	MyProviderAdapter
	watcher *fswatch.Watcher
}

func NewLocal(ctx context.Context, router adapter.Router, factory log.Factory, tag string, options option.LocalProviderOptions) (adapter.Provider, error) {
	if options.Path == "" {
		return nil, E.New("path is missing")
	}

	logger := factory.NewLogger(F.ToString("provider/", options.Type, "[", tag, "]"))

	provider := &Local{
		MyProviderAdapter: MyProviderAdapter{
			Adapter: provider.NewAdapter(C.TypeLocal, tag),
			ctx:     ctx,
			logger:  logger,

			path: options.Path,
		},
	}

	process, err := NewProcessOptions(options.Filter)
	if err != nil {
		return nil, err
	}
	provider.process = process
	return provider, nil
}

func (l *Local) parseProviderFile() error {
	fileInfo, err := os.Stat(l.path)
	if err != nil {
		return err
	}
	l.lastUpdateTime = fileInfo.ModTime()
	content, err := os.ReadFile(l.path)
	if err != nil {
		return err
	}

	l.logger.DebugContext(l.ctx, "parser raw config")
	options, err := NewParser(l.ctx, content)
	if err != nil {
		l.logger.ErrorContext(l.ctx, "parser raw config failed: ", err)
		return err
	}

	l.UpdateOutbounds(l.lastOutOpts, options.Outbounds)
	return nil
}

func (l *Local) Start(stage adapter.StartStage) error {
	switch stage {
	case adapter.StartStateInitialize:
		err := l.parseProviderFile()
		if err != nil {
			return err
		}
	case adapter.StartStatePostStart:
		filePath, _ := filepath.Abs(l.path)
		watcher, err := fswatch.NewWatcher(fswatch.Options{
			Path: []string{filePath},
			Callback: func(path string) {
				err := l.UpdateProvider()
				if err != nil {
					l.logger.Error(err, "reload provider ")
				}
			},
		})
		if err != nil {
			return err
		}
		l.watcher = watcher

		go l.loopUpdate()
	}
	return nil
}

func (s *Local) loopUpdate() error {
	if s.watcher != nil {
		err := s.watcher.Start()
		if err != nil {
			s.logger.Error(err)
			return err
		}
	}
	return nil
}

func (l *Local) Close() error {
	return common.Close(common.PtrOrNil(l.watcher))
}
