package provider

import (
	"encoding/base64"
	"strings"

	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/provider/manager"
	E "github.com/sagernet/sing/common/exceptions"
)

// NewRawParser attempts to decode the given content as URL-safe Base64;
// if successful, it parses the decoded subscription. Otherwise, it treats
// the original content as raw subscription data.
func NewRawParser(content string) (*manager.Options, error) {
	decoded, err := tryDecodeURLSafeBase64(content)
	if err == nil {
		if opts, err := parseRawSubscription(decoded); err == nil && len(opts.Outbounds) > 0 {
			return opts, nil
		}
	}
	return parseRawSubscription(content)
}

// tryDecodeURLSafeBase64 decodes URL-safe Base64, normalizing common
// URL-safe characters. Returns an error if decoding fails.
func tryDecodeURLSafeBase64(input string) (string, error) {
	// Normalize URL-safe characters
	replacer := strings.NewReplacer(
		" ", "=",
		"-", "+",
		"_", "/",
	)
	norm := replacer.Replace(input)

	// Pad to proper length for standard Base64
	if m := len(norm) % 4; m != 0 {
		norm += strings.Repeat("=", 4-m)
	}

	bytes, err := base64.StdEncoding.DecodeString(norm)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// parseRawSubscription splits the subscription content by newlines and
// collects valid Outbound configurations. Returns an error if none found.
func parseRawSubscription(raw string) (*manager.Options, error) {
	lines := strings.FieldsFunc(raw, func(r rune) bool {
		// split on any linebreak
		return r == '\r' || r == '\n'
	})

	var outbounds []option.Outbound
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		server, err := ParseSubscriptionLink(line)
		if err != nil {
			continue // ignore invalid links
		}
		outbounds = append(outbounds, *server)
	}

	if len(outbounds) == 0 {
		return nil, E.New("no servers found in subscription content")
	}

	return &manager.Options{
		Type: C.TypeRawConfig,
		Options: option.Options{
			Outbounds: outbounds,
		},
	}, nil
}
