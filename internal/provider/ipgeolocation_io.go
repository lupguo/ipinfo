package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lupguo/ip_info/internal/model"
)

// IPGeolocationIO queries api.ipgeolocation.io (requires API key).
type IPGeolocationIO struct {
	baseURL string
	token   string
}

func NewIPGeolocationIO(baseURL, token string) *IPGeolocationIO {
	return &IPGeolocationIO{baseURL: strings.TrimRight(baseURL, "/"), token: token}
}

func (p *IPGeolocationIO) Name() string { return "ipgeolocation.io" }

func (p *IPGeolocationIO) Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error) {
	if p.token == "" {
		return nil, fmt.Errorf("ipgeolocation.io: API key required")
	}

	url := p.baseURL + "?apiKey=" + p.token
	if ip != "" {
		url += "&ip=" + ip
	}

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
		return nil, fmt.Errorf("ipgeolocation.io: HTTP %d", resp.StatusCode)
	}

	var raw struct {
		IP              string `json:"ip"`
		CountryName     string `json:"country_name"`
		StateProv       string `json:"state_prov"`
		City            string `json:"city"`
		Zipcode         string `json:"zipcode"`
		Latitude        string `json:"latitude"`
		Longitude       string `json:"longitude"`
		TimezoneName    string `json:"time_zone"`
		Organization    string `json:"organization"`
		ISP             string `json:"isp"`
		ConnectionType  string `json:"connection_type"`
		TimeZone        struct {
			Name string `json:"name"`
		} `json:"time_zone_struct"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Message != "" {
		return nil, fmt.Errorf("ipgeolocation.io: %s", raw.Message)
	}

	lat, _ := strconv.ParseFloat(raw.Latitude, 64)
	lon, _ := strconv.ParseFloat(raw.Longitude, 64)

	return &model.IPInfo{
		IP:       raw.IP,
		Country:  raw.CountryName,
		Region:   raw.StateProv,
		City:     raw.City,
		Org:      raw.Organization,
		Lat:      lat,
		Lon:      lon,
		Zip:      raw.Zipcode,
		Timezone: raw.TimezoneName,
		ISP:      raw.ISP,
		Type:     raw.ConnectionType,
		Source:   p.Name(),
	}, nil
}
