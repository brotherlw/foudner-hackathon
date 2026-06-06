package gateway

import (
	"net/http"
	"strings"
	"time"
)

type ResourceConfig struct {
	Path        string
	Amount      string
	Currency    string
	Description string
}

type PaywallConfig struct {
	BaseURL   string
	ChallengeTTL time.Duration
}

func PaywallMiddleware(cfg PaywallConfig, resource ResourceConfig, grants *GrantStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		grantToken := extractGrantToken(r)
		if grantToken != "" {
			if _, err := grants.VerifyGrant(grantToken, resource.Path); err == nil {
				next.ServeHTTP(w, r)
				return
			}
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
		_, _ = w.Write(body)
	})
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
