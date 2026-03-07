package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lupguo/ip_info/internal/model"
)

// IPInfoIO queries ipinfo.io.
type IPInfoIO struct {
	baseURL string
	token   string
}

func NewIPInfoIO(baseURL, token string) *IPInfoIO {
	return &IPInfoIO{baseURL: strings.TrimRight(baseURL, "/"), token: token}
}

func (p *IPInfoIO) Name() string { return "ipinfo.io" }

func (p *IPInfoIO) Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error) {
	url := p.baseURL + "/"
	if ip != "" {
		url += ip + "/"
	}
	url += "json"
	if p.token != "" {
		url += "?token=" + p.token
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
		return nil, fmt.Errorf("ipinfo.io: HTTP %d", resp.StatusCode)
	}

	var raw struct {
		IP       string `json:"ip"`
		City     string `json:"city"`
		Region   string `json:"region"`
		Country  string `json:"country"`
		Loc      string `json:"loc"` // "lat,lon"
		Org      string `json:"org"`
		Postal   string `json:"postal"`
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	info := &model.IPInfo{
		IP:       raw.IP,
		Country:  raw.Country,
		Region:   raw.Region,
		City:     raw.City,
		Org:      raw.Org,
		Zip:      raw.Postal,
		Timezone: raw.Timezone,
		Source:   p.Name(),
	}
	// Parse "lat,lon"
	fmt.Sscanf(raw.Loc, "%f,%f", &info.Lat, &info.Lon)
	return info, nil
}
