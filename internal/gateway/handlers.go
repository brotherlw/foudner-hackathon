package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type GrantVerifyHandler struct {
	Grants *GrantStore
}

type GrantVerifyResponse struct {
	Ready       bool   `json:"ready"`
	AccessGrant string `json:"access_grant,omitempty"`
}

func (h *GrantVerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	paymentID := r.URL.Query().Get("payment_id")
	if paymentID == "" {
		http.Error(w, "payment_id required", http.StatusBadRequest)
		return
	}
	grant, ok := h.Grants.GetPendingGrant(paymentID)
	resp := GrantVerifyResponse{Ready: ok}
	if ok {
		resp.AccessGrant = grant
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func PremiumReportHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.Copy(w, bytes.NewBufferString(premiumReportContent()))
	})
}

func premiumReportContent() string {
	return `Premium Market Summary — June 2026

EUR/USD: 1.08 (+0.2%)
DAX: 18,420 (+1.1%)
STOXX 600: 518 (+0.8%)

Sector highlights:
- Technology leads gains on AI infrastructure demand
- Energy mixed amid stable Brent at €78/bbl
- Financials supported by rate-cut expectations

This report is available only to paid agent clients.
`
}
