package gateway

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/ledger"
	"github.com/agentic-paywall/agentic-paywall/internal/payments"
)

var ErrWebhookIgnored = errors.New("webhook ignored")

type Server struct {
	cfg      *config.Config
	grants   *GrantStore
	ledger   ledger.Ledger
	provider payments.PaymentProvider
	mu       sync.Mutex
	payments map[string]paymentMeta
}

type paymentMeta struct {
	ResourcePath string
	Amount       string
	Currency     string
}

func NewServer(cfg *config.Config, provider payments.PaymentProvider, ledgers ...ledger.Ledger) *Server {
	ttl := time.Duration(cfg.Gateway.GrantTTLSeconds) * time.Second
	l := ledger.Ledger(ledger.NopLedger{})
	if len(ledgers) > 0 && ledgers[0] != nil {
		l = ledgers[0]
	}
	return &Server{
		cfg:      cfg,
		grants:   NewGrantStore(cfg.Gateway.GrantSecret, ttl),
		ledger:   l,
		provider: provider,
		payments: make(map[string]paymentMeta),
	}
}

func (s *Server) Grants() *GrantStore {
	return s.grants
}

func (s *Server) InitiatePayment(ctx context.Context, req InitiateRequest) (InitiateResponse, error) {
	payment, err := s.provider.CreatePayment(ctx, payments.CreatePaymentRequest{
		ResourcePath: req.ResourcePath,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Description:  req.Description,
		WebhookURL:   s.cfg.Gateway.BaseURL + "/webhooks/payment",
	})
	if err != nil {
		return InitiateResponse{}, err
	}

	s.mu.Lock()
	s.payments[payment.ID] = paymentMeta{
		ResourcePath: req.ResourcePath,
		Amount:       req.Amount,
		Currency:     req.Currency,
	}
	s.mu.Unlock()

	appendLedger(s.ledger, ledger.Event{
		Type:         "payment_initiated",
		PaymentID:    payment.ID,
		ResourcePath: req.ResourcePath,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Decision:     string(payment.Status),
	})

	return InitiateResponse{
		PaymentID: payment.ID,
		Status:    string(payment.Status),
	}, nil
}

func (s *Server) HandleWebhookPaid(ctx context.Context, payload WebhookPayload) error {
	if _, ok := s.grants.GetPendingGrant(payload.PaymentID); ok {
		return nil
	}

	payment, err := s.provider.GetPayment(ctx, payload.PaymentID)
	if err != nil {
		return ErrWebhookIgnored
	}
	if payment.Status != payments.StatusPaid {
		return ErrWebhookIgnored
	}

	meta, ok := s.lookupPayment(payload.PaymentID)
	if !ok {
		meta = paymentMeta{}
	}
	if meta.ResourcePath == "" {
		meta.ResourcePath = payment.ResourcePath
	}
	if meta.Amount == "" {
		meta.Amount = payment.Amount
	}
	if meta.Currency == "" {
		meta.Currency = payment.Currency
	}
	if meta.ResourcePath == "" {
		meta.ResourcePath = payload.ResourcePath
	}
	if meta.Amount == "" {
		meta.Amount = payload.Amount
	}
	if meta.Currency == "" {
		meta.Currency = payload.Currency
	}
	if meta.ResourcePath == "" || meta.Amount == "" || meta.Currency == "" {
		return fmt.Errorf("missing payment metadata for %s", payload.PaymentID)
	}
	appendLedger(s.ledger, ledger.Event{
		Type:         "payment_paid",
		PaymentID:    payment.ID,
		ResourcePath: meta.ResourcePath,
		Amount:       meta.Amount,
		Currency:     meta.Currency,
		Decision:     "paid",
	})

	grant, err := s.grants.IssueGrant(meta.ResourcePath, payload.PaymentID, meta.Amount, meta.Currency)
	if err != nil {
		return err
	}
	s.grants.StorePendingGrant(payload.PaymentID, grant)
	appendLedger(s.ledger, ledger.Event{
		Type:         "grant_issued",
		PaymentID:    payload.PaymentID,
		ResourcePath: meta.ResourcePath,
		Amount:       meta.Amount,
		Currency:     meta.Currency,
		Decision:     "granted",
	})
	return nil
}

func (s *Server) lookupPayment(id string) (paymentMeta, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta, ok := s.payments[id]
	return meta, ok
}
