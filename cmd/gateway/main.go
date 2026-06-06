package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/gateway"
	"github.com/agentic-paywall/agentic-paywall/internal/payments"
	"github.com/agentic-paywall/agentic-paywall/internal/payments/setup"
)

func main() {
	log.SetOutput(os.Stderr)

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	provider, err := setup.NewProvider(cfg)
	if err != nil {
		log.Fatalf("provider: %v", err)
	}

	srv := gateway.NewServer(cfg, provider)
	mux := http.NewServeMux()

	challengeTTL := 15 * time.Minute
	paywallCfg := gateway.PaywallConfig{
		BaseURL:      cfg.Gateway.BaseURL,
		ChallengeTTL: challengeTTL,
	}

	for _, res := range cfg.ProtectedResources {
		resource := gateway.ResourceConfig{
			Path:        res.Path,
			Amount:      res.Amount,
			Currency:    res.Currency,
			Description: res.Description,
		}
		var handler http.Handler
		switch res.Path {
		case "/api/premium-report":
			handler = gateway.PremiumReportHandler()
		default:
			handler = http.NotFoundHandler()
		}
		mux.Handle(res.Path, gateway.PaywallMiddleware(paywallCfg, resource, srv.Grants(), handler))
	}

	mux.Handle("/pay/initiate", &gateway.PayInitiateHandler{Initiator: srv})
	mux.Handle("/grants/verify", &gateway.GrantVerifyHandler{Grants: srv.Grants()})
	mux.Handle("/webhooks/payment", &gateway.WebhookHandler{
		Grants: srv.Grants(),
		OnPaid: func(ctx context.Context, payload gateway.WebhookPayload) error {
			return enrichAndHandlePaid(ctx, srv, provider, payload)
		},
	})

	addr := cfg.Gateway.ListenAddr
	log.Printf("gateway listening on %s (provider=%s)", addr, cfg.Provider)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func loadConfig() (*config.Config, error) {
	path := "config.json"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		path = p
	}
	if _, err := os.Stat(path); err == nil {
		return config.Load(path)
	}
	return config.Default(), nil
}

func enrichAndHandlePaid(ctx context.Context, srv *gateway.Server, provider payments.PaymentProvider, payload gateway.WebhookPayload) error {
	if payload.ResourcePath == "" || payload.Amount == "" || payload.Currency == "" {
		payment, err := provider.GetPayment(ctx, payload.PaymentID)
		if err != nil {
			return fmt.Errorf("lookup payment: %w", err)
		}
		if payload.ResourcePath == "" {
			payload.ResourcePath = payment.ResourcePath
		}
		if payload.Amount == "" {
			payload.Amount = payment.Amount
		}
		if payload.Currency == "" {
			payload.Currency = payment.Currency
		}
	}
	return srv.HandleWebhookPaid(ctx, payload)
}
