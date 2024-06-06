package betterjson

import (
	"encoding/json"

	"github.com/hjson/hjson-go/v4"
	"github.com/titanous/json5"
	"gopkg.in/yaml.v3"
)

func convertYamlToJSON(content []byte) ([]byte, error) {
	var mapRaw map[string]any
	if err := yaml.Unmarshal(content, &mapRaw); err != nil {
		return nil, err
	}
	mapClear := make(map[string]any)
	for _, key := range []string{"$schema", "log", "dns", "ntp", "inbounds", "outbounds", "route", "outbound_providers", "experimental"} {
		if value, ok := mapRaw[key]; ok {
			mapClear[key] = value
		}
	}
	return json.Marshal(&mapClear)
}

func convertJson5ToJSON(content []byte) ([]byte, error) {
	var mapRaw map[string]any
	if err := json5.Unmarshal(content, &mapRaw); err != nil {
		return nil, err
	}
	mapClear := make(map[string]any)
	for _, key := range []string{"$schema", "log", "dns", "ntp", "inbounds", "outbounds", "route", "outbound_providers", "experimental"} {
		if value, ok := mapRaw[key]; ok {
			mapClear[key] = value
		}
	}
	return json.Marshal(&mapClear)
}

func convertHjson5ToJSON(content []byte) ([]byte, error) {
	var mapRaw map[string]any
	if err := hjson.Unmarshal(content, &mapRaw); err != nil {
		return nil, err
	}
	mapClear := make(map[string]any)
	for _, key := range []string{"$schema", "log", "dns", "ntp", "inbounds", "outbounds", "route", "outbound_providers", "experimental"} {
		if value, ok := mapRaw[key]; ok {
			mapClear[key] = value
		}
	}
	return json.Marshal(&mapClear)
}

func PreConvert(content []byte) ([]byte, error) {
	if parsedContent, err := convertJson5ToJSON(content); err == nil {
		return parsedContent, nil
	} else if parsedContent, err = convertHjson5ToJSON(content); err == nil {
		return parsedContent, nil
	}
	return convertYamlToJSON(content)
}
