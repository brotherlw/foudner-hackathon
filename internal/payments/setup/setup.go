package setup

import (
	"fmt"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/payments"
	"github.com/agentic-paywall/agentic-paywall/internal/payments/mock"
	mollieprovider "github.com/agentic-paywall/agentic-paywall/internal/payments/mollie"
)

func NewProvider(cfg *config.Config) (payments.PaymentProvider, error) {
	switch cfg.Provider {
	case "mock":
		return mock.NewProvider(cfg.Mollie.WebhookURL), nil
	case "mollie":
		return mollieprovider.NewProvider(cfg.Mollie.APIKeyEnv, cfg.Mollie.WebhookURL, cfg.Mollie.RedirectURL)
	default:
		return nil, fmt.Errorf("unknown payment provider: %s", cfg.Provider)
	}
}
