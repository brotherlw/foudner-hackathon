package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/payments"
)

type fakePaymentProvider struct {
	mu       sync.Mutex
	payments map[string]payments.Payment
	getCalls map[string]int
}

func newFakePaymentProvider(items ...payments.Payment) *fakePaymentProvider {
	p := &fakePaymentProvider{
		payments: make(map[string]payments.Payment),
		getCalls: make(map[string]int),
	}
	for _, item := range items {
		p.payments[item.ID] = item
	}
	return p
}

func (p *fakePaymentProvider) CreatePayment(ctx context.Context, req payments.CreatePaymentRequest) (payments.Payment, error) {
	return payments.Payment{}, errors.New("not implemented")
}

func (p *fakePaymentProvider) GetPayment(ctx context.Context, paymentID string) (payments.Payment, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.getCalls[paymentID]++
	payment, ok := p.payments[paymentID]
	if !ok {
		return payments.Payment{}, errors.New("payment not found")
	}
	return payment, nil
}

func (p *fakePaymentProvider) CompleteTestPayment(ctx context.Context, paymentID string) error {
	return errors.New("not implemented")
}

func (p *fakePaymentProvider) getCallCount(paymentID string) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.getCalls[paymentID]
}

func TestWebhookVerifiesPaymentBeforeGrant(t *testing.T) {
	tests := []struct {
		name       string
		payment    payments.Payment
		wantGrant  bool
		wantStatus string
	}{
		{
			name: "forged unpaid webhook creates no grant",
			payment: payments.Payment{
				ID:           "pay_unpaid",
				Status:       payments.StatusOpen,
				ResourcePath: "/api/premium-report",
				Amount:       "0.50",
				Currency:     "EUR",
			},
			wantGrant:  false,
			wantStatus: `{"ok":true,"ignored":true}`,
		},
		{
			name: "genuine paid webhook creates grant",
			payment: payments.Payment{
				ID:           "pay_paid",
				Status:       payments.StatusPaid,
				ResourcePath: "/api/premium-report",
				Amount:       "0.50",
				Currency:     "EUR",
			},
			wantGrant:  true,
			wantStatus: `{"ok":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := newFakePaymentProvider(tt.payment)
			srv := NewServer(config.Default(), provider)
			handler := &WebhookHandler{Grants: srv.Grants(), OnPaid: srv.HandleWebhookPaid}

			rec := postWebhook(t, handler, WebhookPayload{
				PaymentID: tt.payment.ID,
				Status:    "paid",
			})
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body %q", rec.Code, http.StatusOK, rec.Body.String())
			}
			if body := rec.Body.String(); body != tt.wantStatus {
				t.Fatalf("body = %q, want %q", body, tt.wantStatus)
			}

			grant, ok := srv.Grants().GetPendingGrant(tt.payment.ID)
			if ok != tt.wantGrant {
				t.Fatalf("grant exists = %v, want %v", ok, tt.wantGrant)
			}
			if tt.wantGrant {
				if grant == "" {
					t.Fatal("grant is empty")
				}
				if _, err := srv.Grants().VerifyGrant(grant, tt.payment.ResourcePath); err != nil {
					t.Fatalf("verify grant: %v", err)
				}
			}
			if got := provider.getCallCount(tt.payment.ID); got != 1 {
				t.Fatalf("GetPayment calls = %d, want 1", got)
			}
		})
	}
}

func TestWebhookDuplicatePaymentKeepsSingleGrant(t *testing.T) {
	payment := payments.Payment{
		ID:           "pay_duplicate",
		Status:       payments.StatusPaid,
		ResourcePath: "/api/premium-report",
		Amount:       "0.50",
		Currency:     "EUR",
	}
	provider := newFakePaymentProvider(payment)
	srv := NewServer(config.Default(), provider)
	handler := &WebhookHandler{Grants: srv.Grants(), OnPaid: srv.HandleWebhookPaid}

	first := postWebhook(t, handler, WebhookPayload{PaymentID: payment.ID, Status: "paid"})
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d, want %d; body %q", first.Code, http.StatusOK, first.Body.String())
	}
	firstGrant, ok := srv.Grants().GetPendingGrant(payment.ID)
	if !ok {
		t.Fatal("first webhook did not create grant")
	}

	second := postWebhook(t, handler, WebhookPayload{PaymentID: payment.ID, Status: "paid"})
	if second.Code != http.StatusOK {
		t.Fatalf("second status = %d, want %d; body %q", second.Code, http.StatusOK, second.Body.String())
	}
	secondGrant, ok := srv.Grants().GetPendingGrant(payment.ID)
	if !ok {
		t.Fatal("second webhook removed grant")
	}
	if secondGrant != firstGrant {
		t.Fatal("duplicate webhook replaced existing grant")
	}
	if got := provider.getCallCount(payment.ID); got != 1 {
		t.Fatalf("GetPayment calls = %d, want 1", got)
	}
}

func postWebhook(t *testing.T, handler http.Handler, payload WebhookPayload) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/webhooks/payment", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
