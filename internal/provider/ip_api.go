package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lupguo/ip_info/internal/model"
)

// IPApi queries ip-api.com (free, no token required).
type IPApi struct {
	baseURL string
}

func NewIPApi(baseURL string) *IPApi {
	return &IPApi{baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *IPApi) Name() string { return "ip-api.com" }

func (p *IPApi) Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error) {
	url := p.baseURL
	if ip != "" {
		url += "/" + ip
	}
	// Request extra fields
	url += "?fields=status,message,country,countryCode,regionName,city,zip,lat,lon,timezone,isp,org,as,query"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip-api.com: HTTP %d", resp.StatusCode)
	}

	var raw struct {
		Status     string  `json:"status"`
		Message    string  `json:"message"`
		Query      string  `json:"query"`
		Country    string  `json:"country"`
		Region     string  `json:"regionName"`
		City       string  `json:"city"`
		Zip        string  `json:"zip"`
		Lat        float64 `json:"lat"`
		Lon        float64 `json:"lon"`
		Timezone   string  `json:"timezone"`
		ISP        string  `json:"isp"`
		Org        string  `json:"org"`
		AS         string  `json:"as"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Status != "success" {
		return nil, fmt.Errorf("ip-api.com: %s", raw.Message)
	}

	return &model.IPInfo{
		IP:       raw.Query,
		Country:  raw.Country,
		Region:   raw.Region,
		City:     raw.City,
		Org:      raw.Org,
		Lat:      raw.Lat,
		Lon:      raw.Lon,
		Zip:      raw.Zip,
		Timezone: raw.Timezone,
		ISP:      raw.ISP,
		ASN:      raw.AS,
		Source:   p.Name(),
	}, nil
}
