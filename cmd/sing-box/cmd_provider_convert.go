package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/provider"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/rw"

	"github.com/spf13/cobra"
)

var flagProviderConvertOutput string
var flagProviderConvertUA string

var commandProviderConvert = &cobra.Command{
	Use:   "convert <url>",
	Short: "Convert to sing-box config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := convertProvider(args[0])
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	commandProviderConvert.Flags().StringVarP(&flagProviderConvertOutput, "output", "o", "provider.json", "Output file")
	commandProviderConvert.Flags().StringVarP(&flagProviderConvertUA, "ua", "a", "sing-box", "User-Agent")
	commandProvider.AddCommand(commandProviderConvert)
}

func convertProvider(sourceUrl string) error {
	var parsedURL *url.URL
	parsedURL, err := url.Parse(sourceUrl)
	if err != nil {
		return err
	}
	switch parsedURL.Scheme {
	case "":
		parsedURL.Scheme = "http"
	case "http", "https":
	default:
		return E.New("unsupported scheme: ", parsedURL.Scheme)
	}

	instance, err := createPreStartedClient()
	if err != nil {
		return err
	}
	defer instance.Close()
	httpClient = &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialer, err := createDialer(instance, commandToolsFlagOutbound)
				if err != nil {
					return nil, err
				}
				return dialer.DialContext(ctx, network, M.ParseSocksaddr(addr))
			},
			ForceAttemptHTTP2: true,
		},
	}
	defer httpClient.CloseIdleConnections()

	request, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return err
	}
	request.Header.Add("User-Agent", flagProviderConvertUA)
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		response.Body.Close()
		return err
	}

	ctx := box.Context(context.Background(), nil, include.OutboundRegistry(), nil, nil, nil, nil)
	options, err := provider.NewParser(ctx, content)
	if err != nil {
		response.Body.Close()
		return err
	}
	response.Body.Close()

	var outbounds []json.RawMessage
	if options.Outbounds != nil {
		for _, outbound := range options.Outbounds {
			data, err := outbound.MarshalJSONContext(context.Background())
			if err != nil {
				return err
			}
			outbounds = append(outbounds, data)
		}
	}

	cacheContentMap := make(map[string]any)
	// cacheContentMap["type"] = options.Type
	cacheContentMap["outbounds"] = outbounds

	content, err = json.Marshal(cacheContentMap)
	if err != nil {
		return err
	}
	err = rw.MkdirParent(flagProviderConvertOutput)
	if err != nil {
		return err
	}
	err = os.WriteFile(flagProviderConvertOutput, content, 0o644)
	if err != nil {
		return err
	}
	return err
}
