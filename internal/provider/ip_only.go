package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lupguo/ip_info/internal/model"
)

// IPOnly is a shared adapter for providers that return only a raw IP string
// (e.g. api.ipify.org, icanhazip.com, checkip.amazonaws.com).
type IPOnly struct {
	name    string
	baseURL string
}

func NewIPOnly(name, baseURL string) *IPOnly {
	return &IPOnly{name: name, baseURL: strings.TrimRight(baseURL, "/")}
}

func (p *IPOnly) Name() string { return p.name }

func (p *IPOnly) Query(ctx context.Context, client *http.Client, _ string) (*model.IPInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %d", p.name, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ip := strings.TrimSpace(string(body))
	if ip == "" {
		return nil, fmt.Errorf("%s: empty response", p.name)
	}

	return &model.IPInfo{
		IP:     ip,
		Source: p.Name(),
	}, nil
}
