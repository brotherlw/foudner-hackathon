package gateway

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/ledger"
)

type ResourceConfig struct {
	Path        string
	Amount      string
	Currency    string
	Description string
}

type PaywallConfig struct {
	BaseURL      string
	ChallengeTTL time.Duration
	Ledger       ledger.Ledger
}

func PaywallMiddleware(cfg PaywallConfig, resource ResourceConfig, grants *GrantStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grantToken := extractGrantToken(r)
		if grantToken != "" {
			if claims, err := grants.ConsumeGrant(grantToken, resource.Path); err == nil {
				log.Printf("gateway access_granted resource=%s payment_id=%s amount=%s currency=%s quota=%d", resource.Path, claims.PaymentID, claims.Amount, claims.Currency, claims.Quota)
				appendLedger(cfg.Ledger, ledger.Event{
					Type:         "access_granted",
					ResourcePath: resource.Path,
					Amount:       resource.Amount,
					Currency:     resource.Currency,
					AgentID:      r.Header.Get("AGENT-ID"),
					Decision:     "granted",
				})
				next.ServeHTTP(w, r)
				return
			} else {
				log.Printf("gateway access_denied resource=%s amount=%s currency=%s reason=%s", resource.Path, resource.Amount, resource.Currency, err)
			}
			appendLedger(cfg.Ledger, ledger.Event{
				Type:         "access_denied",
				ResourcePath: resource.Path,
				Amount:       resource.Amount,
				Currency:     resource.Currency,
				AgentID:      r.Header.Get("AGENT-ID"),
				Decision:     "denied",
			})
		}

		challenge := BuildChallenge(
			cfg.BaseURL,
			resource.Path,
			r.Method,
			resource.Description,
			resource.Amount,
			resource.Currency,
			cfg.ChallengeTTL,
		)
		body, err := challenge.JSON()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		headerVal, err := challenge.Base64Header()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("PAYMENT-REQUIRED", headerVal)
		w.WriteHeader(http.StatusPaymentRequired)
		log.Printf("gateway challenge_issued resource=%s method=%s amount=%s currency=%s", resource.Path, r.Method, resource.Amount, resource.Currency)
		appendLedger(cfg.Ledger, ledger.Event{
			Type:         "challenge_issued",
			ResourcePath: resource.Path,
			Amount:       resource.Amount,
			Currency:     resource.Currency,
			AgentID:      r.Header.Get("AGENT-ID"),
			Decision:     "denied",
		})
		_, _ = w.Write(body)
	})
}

func appendLedger(l ledger.Ledger, event ledger.Event) {
	if l == nil {
		return
	}
	_ = l.Append(event)
}

func extractGrantToken(r *http.Request) string {
	if h := r.Header.Get("PAYMENT-GRANT"); h != "" {
		return strings.TrimSpace(h)
	}
	if q := r.URL.Query().Get("access_grant"); q != "" {
		return strings.TrimSpace(q)
	}
	return ""
}
