package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeCompleter struct {
	paymentID string
	err       error
}

func (c *fakeCompleter) CompleteTestPayment(ctx context.Context, paymentID string) error {
	c.paymentID = paymentID
	return c.err
}

func TestCompleteTestPaymentHandler(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		paymentID  string
		err        error
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "disabled",
			enabled:    false,
			paymentID:  "pay_123",
			wantStatus: http.StatusNotFound,
			wantCalled: false,
		},
		{
			name:       "missing payment id",
			enabled:    true,
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
		{
			name:       "completion error",
			enabled:    true,
			paymentID:  "pay_123",
			err:        errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantCalled: true,
		},
		{
			name:       "complete",
			enabled:    true,
			paymentID:  "pay_123",
			wantStatus: http.StatusOK,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completer := &fakeCompleter{err: tt.err}
			handler := &CompleteTestPaymentHandler{Completer: completer, Enabled: tt.enabled}
			body, err := json.Marshal(CompleteTestPaymentRequest{PaymentID: tt.paymentID})
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, "/pay/complete-test", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body %q", rec.Code, tt.wantStatus, rec.Body.String())
			}
			called := completer.paymentID != ""
			if called != tt.wantCalled {
				t.Fatalf("called = %v, want %v", called, tt.wantCalled)
			}
			if called && completer.paymentID != tt.paymentID {
				t.Fatalf("paymentID = %q, want %q", completer.paymentID, tt.paymentID)
			}
		})
	}
}
