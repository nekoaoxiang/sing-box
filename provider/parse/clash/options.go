package provider

import (
	E "github.com/sagernet/sing/common/exceptions"
	"gopkg.in/yaml.v3"
)

type Clash struct {
	Proxies []Proxy `yaml:"proxies"`
}

type _Proxy struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Options any
}

type Proxy _Proxy

func (p *Proxy) UnmarshalYAML(value *yaml.Node) error {
	if err := value.Decode((*_Proxy)(p)); err != nil {
		return err
	}
	options, loaded := globalRegistry.CreateOptions(p.Type)
	if !loaded {
		return E.New("unknown proxy type: ", p.Type)
	}
	if err := value.Decode(options); err != nil {
		return err
	}
	p.Options = options
	return nil
}

type BaseProxy struct {
	Server string    `yaml:"server"`
	Port   uint16    `yaml:"port"`
	UDP    *bool     `yaml:"udp,omitempty"`
	Smux   *smuxOpts `yaml:"smux,omitempty"`
}
