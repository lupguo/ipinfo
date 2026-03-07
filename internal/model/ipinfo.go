package model

// IPInfo holds the result from an IP information provider.
type IPInfo struct {
	// Compact fields — always shown
	IP       string  `json:"ip"`
	Country  string  `json:"country"`
	Region   string  `json:"region"`
	City     string  `json:"city"`
	Org      string  `json:"org"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Zip      string  `json:"zip"`
	Timezone string  `json:"timezone"`

	// Detail-only fields — omit when empty
	ASN     string `json:"asn,omitempty"`
	ISP     string `json:"isp,omitempty"`
	Type    string `json:"type,omitempty"`    // residential / datacenter / mobile
	IsVPN   bool   `json:"is_vpn,omitempty"`
	IsProxy bool   `json:"is_proxy,omitempty"`
	IsTor   bool   `json:"is_tor,omitempty"`
	Source  string `json:"source,omitempty"` // which provider answered
}
