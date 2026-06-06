package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type WebhookPayload struct {
	PaymentID    string `json:"payment_id"`
	Status       string `json:"status"`
	ResourcePath string `json:"resource_path"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
}

type WebhookHandler struct {
	Grants *GrantStore
	OnPaid func(ctx context.Context, payload WebhookPayload) error
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/x-www-form-urlencoded" || r.FormValue("id") != "" {
		_ = r.ParseForm()
		payload.PaymentID = r.FormValue("id")
		payload.Status = "paid"
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
	}

	if payload.PaymentID == "" {
		http.Error(w, "payment_id required", http.StatusBadRequest)
		return
	}
	if payload.Status == "" {
		payload.Status = "paid"
	}
	if payload.Status != "paid" {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"ignored":true}`))
		return
	}
	if h.OnPaid != nil {
		if err := h.OnPaid(r.Context(), payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"ok":true}`))
}
