package gateway

import (
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
			if _, err := grants.VerifyGrant(grantToken, resource.Path); err == nil {
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
