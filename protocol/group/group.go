package group

import (
	"regexp"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
)

type myGroupAdapter struct {
	defaultTags []string
	outbound    adapter.OutboundManager

	tags      []string
	outbounds map[string]adapter.Outbound

	uses                []string
	includeAllProviders bool
	includes            *FilterList
	excludes            *FilterList
	providerManager     adapter.ProviderManager
}

func (g *myGroupAdapter) getOutbounds() error {
	outbounds := make(map[string]adapter.Outbound)
	tags := []string{}
	uses := []string{}
	providers := []adapter.Provider{}
	for i, tag := range g.defaultTags {
		detour, loaded := g.outbound.Outbound(tag)
		if !loaded {
			return E.New("outbound ", i, " not found: ", tag)
		}
		outbounds[tag] = detour
		tags = append(tags, tag)
	}

	if g.includeAllProviders {
		for _, provider := range g.providerManager.Providers() {
			g.uses = append(uses, provider.Tag())
		}
	}
	for i, use := range g.uses {
		provider, loaded := g.providerManager.Provider(use)
		if !loaded {
			return E.New("provider ", i, " not found: ", provider.Tag())
		}
		providers = append(providers, provider)
	}
	for _, provider := range providers {
		if manager := provider.Outbound(); manager != nil {
			for _, outbound := range manager.Outbounds() {
				if !matchProviderFilter(outbound, g.includes, g.excludes) {
					continue
				}
				tag := outbound.Tag()
				outbounds[tag] = outbound
				tags = append(tags, tag)
			}
		}
	}
	if len(outbounds) == 0 {
		direct, _ := g.outbound.Outbound("direct")
		outbounds[direct.Tag()] = direct
		tags = append(tags, direct.Tag())
	}
	g.outbounds = outbounds
	g.tags = tags
	return nil
}

type FilterList struct {
	Regex []*regexp.Regexp
	Types []string
}

func NewProviderFilter(options *option.FilterList) (*FilterList, error) {
	Regexs := make([]*regexp.Regexp, 0, len(options.Regex))
	for i, regex := range options.Regex {
		regex, err := regexp.Compile(regex)
		if err != nil {
			return nil, E.Cause(err, "parse includes[", i, "]")
		}
		Regexs = append(Regexs, regex)
	}
	return &FilterList{
		Regex: Regexs,
		Types: options.Types,
	}, nil
}

func matchProviderFilter(outbound adapter.Outbound, includes *FilterList, excludes *FilterList) bool {
	return TestRegexpMatch(outbound, includes) && TestRegexpMatch(outbound, excludes) && TestTpyeMatch(outbound, includes) && TestTpyeMatch(outbound, excludes)
}

func TestRegexpMatch(outbound adapter.Outbound, filterList *FilterList) bool {
	if filterList != nil {
		if filterList.Regex != nil {
			tag := outbound.Tag()
			Regex := filterList.Regex
			return common.All(Regex, func(it *regexp.Regexp) bool {
				return it.MatchString(tag)
			})
		}
	}
	return true

}

func TestTpyeMatch(outbound adapter.Outbound, filterList *FilterList) bool {
	if filterList != nil {
		if filterList.Types != nil {
			otype := outbound.Type()
			Type := filterList.Types
			return common.Any(Type, func(it string) bool {
				return otype == it
			})
		}
	}
	return true
}
