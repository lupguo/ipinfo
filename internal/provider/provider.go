package provider

import (
	"context"
	"net/http"

	"github.com/lupguo/ip_info/internal/model"
)

// Provider is the common interface every IP provider adapter must implement.
type Provider interface {
	// Name returns the human-readable provider name (matches config).
	Name() string
	// Query fetches IP information using the supplied HTTP client.
	// ip may be empty string, meaning "query the exit IP of this client".
	Query(ctx context.Context, client *http.Client, ip string) (*model.IPInfo, error)
}
