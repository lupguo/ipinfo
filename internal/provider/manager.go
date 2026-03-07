package provider

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/lupguo/ip_info/internal/cache"
	"github.com/lupguo/ip_info/internal/circuit"
	"github.com/lupguo/ip_info/internal/config"
	"github.com/lupguo/ip_info/internal/model"
)

// entry pairs a Provider with its configured priority.
type entry struct {
	p        Provider
	priority int
}

// Manager routes queries across tiered providers with circuit breaking.
type Manager struct {
	entries  []entry
	breakers map[string]*circuit.Breaker
	cache    *cache.Cache
}

// NewManager constructs a Manager from config, wiring up all providers.
func NewManager(cfg *config.Config) *Manager {
	cbCfg := cfg.CircuitBreaker
	breakers := make(map[string]*circuit.Breaker)
	var entries []entry

	for _, pc := range cfg.Providers {
		if !pc.Enabled {
			continue
		}
		p := buildProvider(pc)
		if p == nil {
			continue
		}
		entries = append(entries, entry{p: p, priority: pc.Priority})
		breakers[pc.Name] = circuit.New(cbCfg.MaxFailures, cbCfg.ResetTimeout)
	}

	return &Manager{
		entries:  entries,
		breakers: breakers,
		cache:    cache.New(cfg.Cache.TTL, cfg.Cache.MaxSize),
	}
}

// buildProvider maps a ProviderConfig to a concrete Provider implementation.
func buildProvider(pc config.ProviderConfig) Provider {
	switch pc.Name {
	case "ipinfo.io":
		return NewIPInfoIO(pc.BaseURL, pc.Token)
	case "ip-api.com":
		return NewIPApi(pc.BaseURL)
	case "ipapi.co":
		return NewIPApiCo(pc.BaseURL)
	case "ipgeolocation.io":
		return NewIPGeolocationIO(pc.BaseURL, pc.Token)
	case "ipwho.is":
		return NewIPWhoIs(pc.BaseURL)
	case "ipstack.com":
		return NewIPStackCom(pc.BaseURL, pc.Token)
	case "api.ipify.org", "icanhazip.com", "checkip.amazonaws.com":
		return NewIPOnly(pc.Name, pc.BaseURL)
	default:
		return nil
	}
}

// Query returns IP information for ip. If ip is empty, the client's exit IP is queried.
func (m *Manager) Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error) {
	cacheKey := ip
	if cacheKey == "" {
		cacheKey = "self"
	}

	// 1. Cache hit
	if info, ok := m.cache.Get(cacheKey); ok {
		return info, nil
	}

	// 2. Group providers into tiers (ascending priority order)
	tiers := m.groupByTier()

	// 3. Try each tier in ascending priority order
	for _, tier := range tiers {
		// Filter out open circuit breakers
		var available []Provider
		for _, p := range tier {
			if b := m.breakers[p.Name()]; b != nil && b.IsOpen() {
				continue
			}
			available = append(available, p)
		}
		if len(available) == 0 {
			continue
		}

		// Random shuffle within the tier for load distribution
		rand.Shuffle(len(available), func(i, j int) {
			available[i], available[j] = available[j], available[i]
		})

		for _, p := range available {
			b := m.breakers[p.Name()]
			if b != nil && !b.Allow() {
				continue
			}

			tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			info, err := p.Query(tCtx, client, ip)
			cancel()

			if err != nil {
				if b != nil {
					b.RecordFailure()
				}
				continue
			}

			if b != nil {
				b.RecordSuccess()
			}
			m.cache.Set(cacheKey, info)
			return info, nil
		}
	}

	return nil, errors.New("all providers failed")
}

// groupByTier returns slices of providers grouped by priority, sorted ascending.
func (m *Manager) groupByTier() [][]Provider {
	tierMap := make(map[int][]Provider)
	var keys []int
	seen := make(map[int]bool)

	for _, e := range m.entries {
		tierMap[e.priority] = append(tierMap[e.priority], e.p)
		if !seen[e.priority] {
			seen[e.priority] = true
			keys = append(keys, e.priority)
		}
	}
	sort.Ints(keys)

	result := make([][]Provider, 0, len(keys))
	for _, k := range keys {
		result = append(result, tierMap[k])
	}
	return result
}
