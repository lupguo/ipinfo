package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lupguo/ip_info/internal/model"
)

// IPWhoIs queries ipwho.is (free, no token).
type IPWhoIs struct {
	baseURL string
}

func NewIPWhoIs(baseURL string) *IPWhoIs {
	return &IPWhoIs{baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *IPWhoIs) Name() string { return "ipwho.is" }

func (p *IPWhoIs) Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error) {
	url := p.baseURL + "/"
	if ip != "" {
		url += ip
	}

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
		return nil, fmt.Errorf("ipwho.is: HTTP %d", resp.StatusCode)
	}

	var raw struct {
		IP        string  `json:"ip"`
		Success   bool    `json:"success"`
		Message   string  `json:"message"`
		Country   string  `json:"country"`
		Region    string  `json:"region"`
		City      string  `json:"city"`
		Postal    string  `json:"postal"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timezone  struct {
			ID string `json:"id"`
		} `json:"timezone"`
		Connection struct {
			Org string `json:"org"`
			ISP string `json:"isp"`
			ASN int    `json:"asn"`
		} `json:"connection"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if !raw.Success {
		return nil, fmt.Errorf("ipwho.is: %s", raw.Message)
	}

	return &model.IPInfo{
		IP:       raw.IP,
		Country:  raw.Country,
		Region:   raw.Region,
		City:     raw.City,
		Org:      raw.Connection.Org,
		Lat:      raw.Latitude,
		Lon:      raw.Longitude,
		Zip:      raw.Postal,
		Timezone: raw.Timezone.ID,
		ISP:      raw.Connection.ISP,
		ASN:      fmt.Sprintf("AS%d", raw.Connection.ASN),
		Type:     raw.Type,
		Source:   p.Name(),
	}, nil
}
