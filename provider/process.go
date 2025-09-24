package provider

import (
	"regexp"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
)

type ProcessOptions struct {
	include []*regexp.Regexp
	exclude []*regexp.Regexp
	Invert  bool
}

func NewProcessOptions(options *option.FilterOptions) (*ProcessOptions, error) {
	if options == nil {
		return nil, nil
	}
	var (
		include []*regexp.Regexp
		exclude []*regexp.Regexp
	)
	if options.Includes != nil {
		for regexIndex, it := range options.Includes.Regex {
			regex, err := regexp.Compile(it)
			if err != nil {
				return nil, E.Cause(err, "parse filter[", regexIndex, "]")
			}
			include = append(include, regex)
		}
	}
	if options.Excludes != nil {
		for regexIndex, it := range options.Excludes.Regex {
			regex, err := regexp.Compile(it)
			if err != nil {
				return nil, E.Cause(err, "parse exclude[", regexIndex, "]")
			}
			exclude = append(exclude, regex)
		}
	}
	return &ProcessOptions{
		include: include,
		exclude: exclude,
		Invert:  options.Invert,
	}, nil
}

func (o *ProcessOptions) Process(tag string) bool {
	if o == nil {
		return true
	}
	inProcess := false
	if len(o.include) > 0 {
		inProcess = common.Any(o.include, func(it *regexp.Regexp) bool {
			return it.MatchString(tag)
		})
	}
	if inProcess && len(o.exclude) > 0 {
		if !common.Any(o.exclude, func(it *regexp.Regexp) bool {
			return it.MatchString(tag)
		}) {
			inProcess = false
		}
	}

	if o.Invert {
		inProcess = !inProcess
	}

	return inProcess
}
