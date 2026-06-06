package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/payments"
)

type Server struct {
	cfg      *config.Config
	grants   *GrantStore
	provider payments.PaymentProvider
	mu       sync.Mutex
	payments map[string]paymentMeta
}

type paymentMeta struct {
	ResourcePath string
	Amount       string
	Currency     string
}

func NewServer(cfg *config.Config, provider payments.PaymentProvider) *Server {
	ttl := time.Duration(cfg.Gateway.GrantTTLSeconds) * time.Second
	return &Server{
		cfg:      cfg,
		grants:   NewGrantStore(cfg.Gateway.GrantSecret, ttl),
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

	return InitiateResponse{
		PaymentID: payment.ID,
		Status:    string(payment.Status),
	}, nil
}

func (s *Server) HandleWebhookPaid(ctx context.Context, payload WebhookPayload) error {
	meta, ok := s.lookupPayment(payload.PaymentID)
	if !ok {
		meta = paymentMeta{
			ResourcePath: payload.ResourcePath,
			Amount:       payload.Amount,
			Currency:     payload.Currency,
		}
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

	grant, err := s.grants.IssueGrant(meta.ResourcePath, payload.PaymentID, meta.Amount, meta.Currency)
	if err != nil {
		return err
	}
	s.grants.StorePendingGrant(payload.PaymentID, grant)
	return nil
}

func (s *Server) lookupPayment(id string) (paymentMeta, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	meta, ok := s.payments[id]
	return meta, ok
}
