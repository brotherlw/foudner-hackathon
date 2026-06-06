package gateway

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type CompleteTestPaymentRequest struct {
	PaymentID string `json:"payment_id"`
}

type CompleteTestPaymentResponse struct {
	OK        bool   `json:"ok"`
	PaymentID string `json:"payment_id"`
}

type TestPaymentCompleter interface {
	CompleteTestPayment(ctx context.Context, paymentID string) error
}

type CompleteTestPaymentHandler struct {
	Completer TestPaymentCompleter
	Enabled   bool
}

func (h *CompleteTestPaymentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.Enabled {
		http.Error(w, "test payment completion disabled", http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var req CompleteTestPaymentRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.PaymentID == "" {
		http.Error(w, "payment_id required", http.StatusBadRequest)
		return
	}
	if err := h.Completer.CompleteTestPayment(r.Context(), req.PaymentID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("gateway complete_test_payment payment_id=%s", req.PaymentID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CompleteTestPaymentResponse{OK: true, PaymentID: req.PaymentID})
}
