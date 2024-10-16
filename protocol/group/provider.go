package group

import (
	"regexp"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
)

type myProviderAdapter struct {
	uses                []string
	includeAllProviders bool
	includes            *FilterList
	excludes            *FilterList
	providerManager     adapter.ProviderManager
}

func (s *myProviderAdapter) getOutbounds(tags []string, outboundManager adapter.OutboundManager) (map[string]adapter.Outbound, []string, error) {
	outbounds := map[string]adapter.Outbound{}
	outboundTags := []string{}

	for i, tag := range tags {
		detour, loaded := outboundManager.Outbound(tag)
		if !loaded {
			return nil, nil, E.New("outbound ", i, " not found: ", tag)
		}
		outbounds[tag] = detour
		outboundTags = append(outboundTags, tag)
	}

	if s.includeAllProviders {
		uses := []string{}
		for _, provider := range s.providerManager.Providers() {
			uses = append(uses, provider.Tag())
		}
		s.uses = uses
	}
	for i, use := range s.uses {
		provider, loaded := s.providerManager.Provider(use)
		if !loaded {
			return nil, nil, E.New("provider ", i, " not found: ", provider)
		}
		for _, outbound := range provider.OutboundManager().Outbounds() {
			if !matchProviderFilter(outbound, s.includes, s.excludes) {
				continue
			}
			tag := outbound.Tag()
			outbounds[tag] = outbound
			outboundTags = append(outboundTags, tag)
		}
	}
	if len(outbounds) == 0 {
		direct, _ := outboundManager.Outbound("direct")
		outbounds[direct.Tag()] = direct
		outboundTags = append(outboundTags, direct.Tag())
	}
	return outbounds, outboundTags, nil
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
