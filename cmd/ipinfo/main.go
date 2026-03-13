package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/proxy"

	"github.com/lupguo/ip_info/internal/config"
	"github.com/lupguo/ip_info/internal/model"
	"github.com/lupguo/ip_info/internal/provider"
)

var version = "dev"

func main() {
	var (
		proxyAddr  string
		noProxy    bool
		detail     bool
		configPath string
		ipFlag     string
	)

	root := &cobra.Command{
		Use:     "ipinfo [IP]",
		Short:   "Query IP geolocation and network information",
		Version: version,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ip := ipFlag
			if ip == "" && len(args) == 1 {
				ip = args[0]
			}

			cfg, err := config.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			client, err := buildHTTPClient(proxyAddr, noProxy)
			if err != nil {
				return fmt.Errorf("building HTTP client: %w", err)
			}

			mgr := provider.NewManager(cfg)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			info, err := mgr.Query(ctx, client, ip)
			if err != nil {
				return fmt.Errorf("query failed: %w", err)
			}

			return printInfo(info, detail)
		},
	}

	root.Flags().StringVarP(&ipFlag, "ip", "i", "", "IP address to query (takes precedence over positional argument)")
	root.Flags().StringVarP(&proxyAddr, "proxy", "x", "", "proxy URL (http://... or socks5h://...)")
	root.Flags().BoolVar(&noProxy, "no-proxy", false, "ignore proxy environment variables (HTTP_PROXY, HTTPS_PROXY, etc.)")
	root.MarkFlagsMutuallyExclusive("proxy", "no-proxy")
	root.Flags().BoolVarP(&detail, "detail", "d", false, "show detailed output")
	root.Flags().StringVarP(&configPath, "config", "c", "", "config file path (default: ~/.ipinfo/config.yaml)")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildHTTPClient constructs an *http.Client, optionally configured with a proxy.
//
// Proxy resolution order:
//  1. --proxy flag: use the specified proxy URL explicitly.
//  2. --no-proxy flag: force direct connections, ignoring environment variables.
//  3. Default (neither flag): honour HTTP_PROXY / HTTPS_PROXY / NO_PROXY env vars.
func buildHTTPClient(proxyAddr string, noProxy bool) (*http.Client, error) {
	transport := &http.Transport{
		// By default, honour environment proxy variables (HTTP_PROXY, HTTPS_PROXY, NO_PROXY).
		Proxy: http.ProxyFromEnvironment,
	}

	if noProxy {
		// Explicitly disable all proxies, including those from environment variables.
		transport.Proxy = func(*http.Request) (*url.URL, error) { return nil, nil }
	} else if proxyAddr != "" {
		u, err := url.Parse(proxyAddr)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", proxyAddr, err)
		}

		switch u.Scheme {
		case "http", "https":
			transport.Proxy = http.ProxyURL(u)
		case "socks5", "socks5h":
			var auth *proxy.Auth
			if u.User != nil {
				pw, _ := u.User.Password()
				auth = &proxy.Auth{
					User:     u.User.Username(),
					Password: pw,
				}
			}
			dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("creating SOCKS5 dialer: %w", err)
			}
			// Use ContextDialer interface if available, otherwise wrap Dial.
			if cd, ok := dialer.(proxy.ContextDialer); ok {
				transport.DialContext = cd.DialContext
			} else {
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}
			}
		default:
			return nil, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
		}
	}

	return &http.Client{Transport: transport}, nil
}

// printInfo outputs IPInfo as pretty JSON, stripping detail fields when not requested.
func printInfo(info *model.IPInfo, detail bool) error {
	var out interface{}
	if detail {
		out = info
	} else {
		out = &compactInfo{
			IP:       info.IP,
			Country:  info.Country,
			Region:   info.Region,
			City:     info.City,
			Org:      info.Org,
			Lat:      info.Lat,
			Lon:      info.Lon,
			Zip:      info.Zip,
			Timezone: info.Timezone,
			Source:   info.Source,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// compactInfo is a JSON-serializable subset of IPInfo without omitempty detail fields.
type compactInfo struct {
	IP       string  `json:"ip"`
	Country  string  `json:"country"`
	Region   string  `json:"region"`
	City     string  `json:"city"`
	Org      string  `json:"org"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Zip      string  `json:"zip"`
	Timezone string  `json:"timezone"`
	Source   string  `json:"source,omitempty"`
}
