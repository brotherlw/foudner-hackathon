package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Currency           string              `json:"currency"`
	Gateway            GatewayConfig       `json:"gateway"`
	Wallet             WalletConfig        `json:"wallet"`
	Provider           string              `json:"provider"`
	Mollie             MollieConfig        `json:"mollie"`
	ProtectedResources []ProtectedResource `json:"protected_resources"`
}

type GatewayConfig struct {
	ListenAddr      string `json:"listen_addr"`
	GrantSecret     string `json:"grant_secret"`
	GrantTTLSeconds int    `json:"grant_ttl_seconds"`
	BaseURL         string `json:"base_url"`
}

type WalletConfig struct {
	DailyBudget        float64 `json:"daily_budget"`
	AutoPayThreshold   float64 `json:"auto_pay_threshold"`
}

type MollieConfig struct {
	APIKeyEnv  string `json:"api_key_env"`
	WebhookURL string `json:"webhook_url"`
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
	return nil
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
