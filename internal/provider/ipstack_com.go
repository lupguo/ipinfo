package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lupguo/ip_info/internal/model"
)

// IPStackCom queries ipstack.com (requires API key).
type IPStackCom struct {
	baseURL string
	token   string
}

func NewIPStackCom(baseURL, token string) *IPStackCom {
	return &IPStackCom{baseURL: strings.TrimRight(baseURL, "/"), token: token}
}

func (p *IPStackCom) Name() string { return "ipstack.com" }

func (p *IPStackCom) Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error) {
	if p.token == "" {
		return nil, fmt.Errorf("ipstack.com: API key required")
	}

	target := "check"
	if ip != "" {
		target = ip
	}
	url := p.baseURL + "/" + target + "?access_key=" + p.token

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
		return nil, fmt.Errorf("ipstack.com: HTTP %d", resp.StatusCode)
	}

	var raw struct {
		IP          string  `json:"ip"`
		CountryName string  `json:"country_name"`
		RegionName  string  `json:"region_name"`
		City        string  `json:"city"`
		Zip         string  `json:"zip"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		TimeZoneID  string  `json:"time_zone_id"`
		// ipstack nests timezone
		TimeZone struct {
			ID string `json:"id"`
		} `json:"time_zone"`
		ASN struct {
			ASN  string `json:"asn"`
			Name string `json:"name"`
		} `json:"asn"`
		Error struct {
			Code int    `json:"code"`
			Info string `json:"info"`
		} `json:"error"`
		Success *bool `json:"success"` // present only on error
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Success != nil && !*raw.Success {
		return nil, fmt.Errorf("ipstack.com: %s", raw.Error.Info)
	}

	tz := raw.TimeZone.ID
	if tz == "" {
		tz = raw.TimeZoneID
	}

	return &model.IPInfo{
		IP:       raw.IP,
		Country:  raw.CountryName,
		Region:   raw.RegionName,
		City:     raw.City,
		Org:      raw.ASN.Name,
		Lat:      raw.Latitude,
		Lon:      raw.Longitude,
		Zip:      raw.Zip,
		Timezone: tz,
		ASN:      raw.ASN.ASN,
		Source:   p.Name(),
	}, nil
}
