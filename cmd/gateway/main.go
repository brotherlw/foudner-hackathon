package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/gateway"
	"github.com/agentic-paywall/agentic-paywall/internal/ledger"
	"github.com/agentic-paywall/agentic-paywall/internal/payments/setup"
	"github.com/agentic-paywall/agentic-paywall/internal/residency"
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

	auditLedger := ledger.NewFileLedger(cfg.Ledger.Path)
	statement := residency.FromConfig(cfg)
	if err := residency.WriteMarkdown("DATA-RESIDENCY-LIVE.md", statement); err != nil {
		log.Fatalf("data residency statement: %v", err)
	}

	srv := gateway.NewServer(cfg, provider, auditLedger)
	mux := http.NewServeMux()

	challengeTTL := 15 * time.Minute
	paywallCfg := gateway.PaywallConfig{
		BaseURL:      cfg.Gateway.BaseURL,
		ChallengeTTL: challengeTTL,
		Ledger:       auditLedger,
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
	mux.Handle("/pay/complete-test", &gateway.CompleteTestPaymentHandler{
		Completer: srv,
		Enabled:   cfg.Demo.EnableTestCompletion,
	})
	mux.Handle("/grants/verify", &gateway.GrantVerifyHandler{Grants: srv.Grants()})
	mux.Handle("/.well-known/agent-paywall-key", &gateway.PublicKeyHandler{Grants: srv.Grants()})
	mux.Handle("/.well-known/data-residency", &gateway.DataResidencyHandler{Statement: statement})
	mux.Handle("/webhooks/payment", &gateway.WebhookHandler{
		Grants: srv.Grants(),
		OnPaid: srv.HandleWebhookPaid,
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
