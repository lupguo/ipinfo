package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lupguo/ip_info/internal/model"
)

// IPApiCo queries ipapi.co.
type IPApiCo struct {
	baseURL string
}

func NewIPApiCo(baseURL string) *IPApiCo {
	return &IPApiCo{baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *IPApiCo) Name() string { return "ipapi.co" }

func (p *IPApiCo) Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error) {
	url := p.baseURL + "/"
	if ip != "" {
		url += ip + "/"
	}
	url += "json/"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipapi.co: HTTP %d", resp.StatusCode)
	}

	var raw struct {
		IP           string  `json:"ip"`
		City         string  `json:"city"`
		Region       string  `json:"region"`
		CountryName  string  `json:"country_name"`
		Postal       string  `json:"postal"`
		Latitude     float64 `json:"latitude"`
		Longitude    float64 `json:"longitude"`
		Timezone     string  `json:"timezone"`
		Org          string  `json:"org"`
		ASN          string  `json:"asn"`
		Error        bool    `json:"error"`
		Reason       string  `json:"reason"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Error {
		return nil, fmt.Errorf("ipapi.co: %s", raw.Reason)
	}

	return &model.IPInfo{
		IP:       raw.IP,
		Country:  raw.CountryName,
		Region:   raw.Region,
		City:     raw.City,
		Org:      raw.Org,
		Lat:      raw.Latitude,
		Lon:      raw.Longitude,
		Zip:      raw.Postal,
		Timezone: raw.Timezone,
		ASN:      raw.ASN,
		Source:   p.Name(),
	}, nil
}
