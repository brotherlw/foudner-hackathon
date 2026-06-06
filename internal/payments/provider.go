package payments

import "context"

type Status string

const (
	StatusOpen   Status = "open"
	StatusPaid   Status = "paid"
	StatusFailed Status = "failed"
)

type CreatePaymentRequest struct {
	ResourcePath   string
	Amount         string
	Currency       string
	Description    string
	IdempotencyKey string
	WebhookURL     string
}

type Payment struct {
	ID           string
	Status       Status
	ResourcePath string
	Amount       string
	Currency     string
	CheckoutURL  string
}

type PaymentProvider interface {
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (Payment, error)
	GetPayment(ctx context.Context, paymentID string) (Payment, error)
	CompleteTestPayment(ctx context.Context, paymentID string) error
}
