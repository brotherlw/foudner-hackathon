package mock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/payments"
	"github.com/google/uuid"
)

type Provider struct {
	webhookURL string
	client     *http.Client
	mu         sync.Mutex
	store      map[string]payments.Payment
}

func NewProvider(webhookURL string) *Provider {
	return &Provider{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
		store:      make(map[string]payments.Payment),
	}
}

func (p *Provider) CreatePayment(ctx context.Context, req payments.CreatePaymentRequest) (payments.Payment, error) {
	id := uuid.NewString()
	payment := payments.Payment{
		ID:           id,
		Status:       payments.StatusPaid,
		ResourcePath: req.ResourcePath,
		Amount:       req.Amount,
		Currency:     req.Currency,
	}

	p.mu.Lock()
	p.store[id] = payment
	p.mu.Unlock()

	webhook := req.WebhookURL
	if webhook == "" {
		webhook = p.webhookURL
	}
	if webhook != "" {
		go p.notifyWebhook(webhook, payment)
	}

	return payment, nil
}

func (p *Provider) GetPayment(ctx context.Context, paymentID string) (payments.Payment, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	payment, ok := p.store[paymentID]
	if !ok {
		return payments.Payment{}, fmt.Errorf("payment not found: %s", paymentID)
	}
	return payment, nil
}

func (p *Provider) CompleteTestPayment(ctx context.Context, paymentID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	payment, ok := p.store[paymentID]
	if !ok {
		return fmt.Errorf("payment not found: %s", paymentID)
	}
	payment.Status = payments.StatusPaid
	p.store[paymentID] = payment
	return nil
}

func (p *Provider) notifyWebhook(webhookURL string, payment payments.Payment) {
	payload := map[string]string{
		"payment_id":    payment.ID,
		"status":        string(payments.StatusPaid),
		"resource_path": payment.ResourcePath,
		"amount":        payment.Amount,
		"currency":      payment.Currency,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
