package option

import "github.com/sagernet/sing/common/json/badoption"

type FilterList struct {
	Regex badoption.Listable[string] `json:"regex,omitempty"`
	Types badoption.Listable[string] `json:"type,omitempty"`
}

type FilterOptions struct {
	Includes *FilterList `json:"include,omitempty"`
	Excludes *FilterList `json:"exclude,omitempty"`
}

type ProviderGroupOptions struct {
	Outbounds           badoption.Listable[string] `json:"outbounds,omitempty"`
	Providers           badoption.Listable[string] `json:"providers,omitempty"`
	IncludeAllProviders bool                       `json:"include_all_providers,omitempty"`
	Filter              *FilterOptions             `json:"filter,omitempty"`
}

type SelectorOutboundOptions struct {
	ProviderGroupOptions
	Default                   string `json:"default,omitempty"`
	InterruptExistConnections bool   `json:"interrupt_exist_connections,omitempty"`
}

type URLTestOutboundOptions struct {
	ProviderGroupOptions
	URL                       string             `json:"url,omitempty"`
	Interval                  badoption.Duration `json:"interval,omitempty"`
	Tolerance                 uint16             `json:"tolerance,omitempty"`
	IdleTimeout               badoption.Duration `json:"idle_timeout,omitempty"`
	InterruptExistConnections bool               `json:"interrupt_exist_connections,omitempty"`
}
