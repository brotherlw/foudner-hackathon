package gateway

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type InitiateRequest struct {
	ResourcePath string `json:"resource_path"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Description  string `json:"description,omitempty"`
}

type InitiateResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

type PaymentInitiator interface {
	InitiatePayment(ctx context.Context, req InitiateRequest) (InitiateResponse, error)
}

type PayInitiateHandler struct {
	Initiator PaymentInitiator
}

func (h *PayInitiateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var req InitiateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.ResourcePath == "" || req.Amount == "" || req.Currency == "" {
		http.Error(w, "resource_path, amount, and currency required", http.StatusBadRequest)
		return
	}

	resp, err := h.Initiator.InitiatePayment(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
