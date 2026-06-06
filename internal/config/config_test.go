package config

import "testing"

func TestValidateDataResidency(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{
			name: "default config is EU safe",
			mutate: func(c *Config) {
				c.Provider = "mock"
			},
			wantErr: false,
		},
		{
			name: "non EU region is rejected",
			mutate: func(c *Config) {
				c.DataResidency.Region = "us-east"
			},
			wantErr: true,
		},
		{
			name: "mollie requires allowed sub processor",
			mutate: func(c *Config) {
				c.Provider = "mollie"
				c.DataResidency.AllowedSubProcessors = []string{"example.eu"}
			},
			wantErr: true,
		},
		{
			name: "localhost webhook is allowed",
			mutate: func(c *Config) {
				c.Mollie.WebhookURL = "http://localhost:3001/webhooks/payment"
			},
			wantErr: false,
		},
		{
			name: "unlisted external webhook is rejected",
			mutate: func(c *Config) {
				c.Mollie.WebhookURL = "https://logs.example.com/webhook"
			},
			wantErr: true,
		},
		{
			name: "listed external webhook is allowed",
			mutate: func(c *Config) {
				c.Mollie.WebhookURL = "https://payments.example.eu/webhook"
				c.DataResidency.AllowedSubProcessors = []string{"mollie.com", "example.eu"}
			},
			wantErr: false,
		},
		{
			name: "unlisted redirect url is rejected",
			mutate: func(c *Config) {
				c.Mollie.RedirectURL = "https://example.com/payment/return"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.mutate(cfg)
			err := cfg.ValidateDataResidency()
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateDataResidency error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
