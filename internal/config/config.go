package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Currency           string              `json:"currency"`
	Gateway            GatewayConfig       `json:"gateway"`
	Wallet             WalletConfig        `json:"wallet"`
	Provider           string              `json:"provider"`
	Mollie             MollieConfig        `json:"mollie"`
	Ledger             LedgerConfig        `json:"ledger"`
	DataResidency      DataResidencyConfig `json:"data_residency"`
	ProtectedResources []ProtectedResource `json:"protected_resources"`
}

type GatewayConfig struct {
	ListenAddr      string `json:"listen_addr"`
	GrantSecret     string `json:"grant_secret"`
	GrantTTLSeconds int    `json:"grant_ttl_seconds"`
	BaseURL         string `json:"base_url"`
}

type WalletConfig struct {
	DailyBudget      float64 `json:"daily_budget"`
	AutoPayThreshold float64 `json:"auto_pay_threshold"`
}

type MollieConfig struct {
	APIKeyEnv  string `json:"api_key_env"`
	WebhookURL string `json:"webhook_url"`
}

type LedgerConfig struct {
	Path string `json:"path"`
}

type DataResidencyConfig struct {
	Region               string   `json:"region"`
	AllowedSubProcessors []string `json:"allowed_sub_processors"`
	IPHashSalt           string   `json:"ip_hash_salt"`
	StorageLocation      string   `json:"storage_location"`
}

type ProtectedResource struct {
	Path        string `json:"path"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyEnv(&cfg)
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("GATEWAY_URL"); v != "" {
		cfg.Gateway.BaseURL = v
	}
	if v := os.Getenv("PAYMENT_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	if v := os.Getenv("GATEWAY_LISTEN_ADDR"); v != "" {
		cfg.Gateway.ListenAddr = v
	}
	if v := os.Getenv("GRANT_SECRET"); v != "" {
		cfg.Gateway.GrantSecret = v
	}
	if v := os.Getenv("GRANT_TTL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Gateway.GrantTTLSeconds = n
		}
	}
	if v := os.Getenv("LEDGER_PATH"); v != "" {
		cfg.Ledger.Path = v
	}
	if v := os.Getenv("DATA_RESIDENCY_REGION"); v != "" {
		cfg.DataResidency.Region = v
	}
	if v := os.Getenv("DATA_RESIDENCY_ALLOWED_SUB_PROCESSORS"); v != "" {
		cfg.DataResidency.AllowedSubProcessors = splitCSV(v)
	}
	if v := os.Getenv("DATA_RESIDENCY_IP_HASH_SALT"); v != "" {
		cfg.DataResidency.IPHashSalt = v
	}
	if v := os.Getenv("DATA_RESIDENCY_STORAGE_LOCATION"); v != "" {
		cfg.DataResidency.StorageLocation = v
	}
}

func (c *Config) validate() error {
	if c.Gateway.ListenAddr == "" {
		c.Gateway.ListenAddr = ":3001"
	}
	if c.Gateway.GrantSecret == "" {
		c.Gateway.GrantSecret = "dev-only-change-me"
	}
	if c.Gateway.GrantTTLSeconds == 0 {
		c.Gateway.GrantTTLSeconds = 3600
	}
	if c.Gateway.BaseURL == "" {
		c.Gateway.BaseURL = "http://localhost:3001"
	}
	if c.Currency == "" {
		c.Currency = "EUR"
	}
	if c.Provider == "" {
		c.Provider = "mock"
	}
	if c.Ledger.Path == "" {
		c.Ledger.Path = "ledger/events.jsonl"
	}
	if c.DataResidency.Region == "" {
		c.DataResidency.Region = "eu"
	}
	if c.DataResidency.StorageLocation == "" {
		c.DataResidency.StorageLocation = "local append-only JSONL ledger"
	}
	if len(c.DataResidency.AllowedSubProcessors) == 0 {
		c.DataResidency.AllowedSubProcessors = []string{"mollie.com"}
	}
	if err := c.ValidateDataResidency(); err != nil {
		return err
	}
	return nil
}

func (c *Config) ValidateDataResidency() error {
	if !strings.HasPrefix(strings.ToLower(c.DataResidency.Region), "eu") {
		return fmt.Errorf("data residency region must be EU, got %q", c.DataResidency.Region)
	}

	allowed := make(map[string]struct{}, len(c.DataResidency.AllowedSubProcessors))
	for _, host := range c.DataResidency.AllowedSubProcessors {
		normalized := normalizeHost(host)
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	if c.Provider == "mollie" && !hostAllowed("mollie.com", allowed) {
		return fmt.Errorf("mollie provider requires mollie.com in data_residency.allowed_sub_processors")
	}
	if err := checkConfiguredEndpoint("mollie.webhook_url", c.Mollie.WebhookURL, allowed); err != nil {
		return err
	}
	return nil
}

func checkConfiguredEndpoint(name, rawURL string, allowed map[string]struct{}) error {
	if rawURL == "" {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("%s must be an absolute URL", name)
	}
	host := normalizeHost(parsed.Hostname())
	if host == "" || isLocalHost(host) {
		return nil
	}
	if hostAllowed(host, allowed) {
		return nil
	}
	return fmt.Errorf("%s host %q is not in data_residency.allowed_sub_processors", name, host)
}

func hostAllowed(host string, allowed map[string]struct{}) bool {
	host = normalizeHost(host)
	for allowedHost := range allowed {
		if host == allowedHost || strings.HasSuffix(host, "."+allowedHost) {
			return true
		}
	}
	return false
}

func isLocalHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimSuffix(host, "/")
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return host
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func Default() *Config {
	return &Config{
		Currency: "EUR",
		Gateway: GatewayConfig{
			ListenAddr:      ":3001",
			GrantSecret:     "dev-only-change-me",
			GrantTTLSeconds: 3600,
			BaseURL:         "http://localhost:3001",
		},
		Wallet: WalletConfig{
			DailyBudget:      5.00,
			AutoPayThreshold: 1.00,
		},
		Provider: "mock",
		Mollie: MollieConfig{
			APIKeyEnv:  "MOLLIE_API_KEY",
			WebhookURL: "http://localhost:3001/webhooks/payment",
		},
		Ledger: LedgerConfig{
			Path: "ledger/events.jsonl",
		},
		DataResidency: DataResidencyConfig{
			Region:               "eu",
			AllowedSubProcessors: []string{"mollie.com"},
			StorageLocation:      "local append-only JSONL ledger",
		},
		ProtectedResources: []ProtectedResource{
			{
				Path:        "/api/premium-report",
				Amount:      "0.50",
				Currency:    "EUR",
				Description: "Premium market summary report",
			},
		},
	}
}
