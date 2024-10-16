package provider

import (
	"context"
	"os"
	"time"

	"github.com/sagernet/fswatch"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/provider"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/protocol/group"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

func RegisterLocal(registry *provider.Registry) {
	provider.Register[option.LocalProviderOptions](registry, C.TypeLocal, NewLocal)
}

var (
	_ adapter.Provider = (*Local)(nil)
)

type Local struct {
	myProviderAdapter
	watcher *fswatch.Watcher
}

func NewLocal(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.LocalProviderOptions) (adapter.Provider, error) {
	healthcheckInterval := time.Duration(options.HealthCheck.Interval)
	healthcheckUrl := options.HealthCheck.Url
	healthcheckEnable := options.HealthCheck.Enable

	if healthcheckInterval == 0 {
		healthcheckInterval = C.DefaultURLTestInterval
	}
	if healthcheckUrl == "" {
		healthcheckUrl = "https://www.gstatic.com/generate_204"
	}
	if !healthcheckEnable {
		healthcheckEnable = false
	}

	if options.Path == "" {
		return nil, E.New("path is missing")
	}

	ctx, cancel := context.WithCancel(ctx)
	provider := &Local{
		myProviderAdapter: myProviderAdapter{
			Adapter:      provider.NewAdapter(C.TypeLocal, tag),
			ctx:          ctx,
			cancel:       cancel,
			path:         options.Path,
			pauseManager: service.FromContext[pause.Manager](ctx),
		},
	}
	if options.Filter.Includes != nil {
		includes, err := group.NewProviderFilter(options.Filter.Includes)
		if err != nil {
			return nil, err
		}
		provider.includes = includes
	}

	if options.Filter.Excludes != nil {
		excludes, err := group.NewProviderFilter(options.Filter.Excludes)
		if err != nil {
			return nil, err
		}
		provider.excludes = excludes
	}

	lastUpdatedTime, err := provider.parseProviderFile()
	if err != nil {
		return nil, err
	}

	provider.lastUpdateTime = lastUpdatedTime

	return provider, nil
}

func (p *Local) parseProviderFile() (time.Time, error) {
	fileInfo, err := os.Stat(p.path)
	if err != nil {
		return time.Time{}, err
	}
	fileModeTime := fileInfo.ModTime()

	content, err := os.ReadFile(p.path)
	if err != nil {
		return time.Time{}, err
	}

	options, err := p.newParser(content)
	if err != nil {
		return time.Time{}, err
	}

	err = p.CreateOutbounds(p.ctx, p.router, options)
	if err != nil {
		return time.Time{}, err
	}
	return fileModeTime, err
}

func (p *Local) PostStart() error {
	watcher, err := fswatch.NewWatcher(fswatch.Options{
		Path: []string{p.path},
		Callback: func(path string) {
			err := p.UpdateProvider(p.ctx)
			if err != nil {
				p.logger.Error(err)
			}
		},
	})
	if err != nil {
		return err
	}
	p.watcher = watcher

	go p.loopUpdate()

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
