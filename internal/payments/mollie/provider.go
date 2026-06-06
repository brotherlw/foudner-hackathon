package mollie

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/payments"
	mollieclient "github.com/mollie/mollie-api-golang"
	"github.com/mollie/mollie-api-golang/models/components"
	"github.com/mollie/mollie-api-golang/models/operations"
)

type Provider struct {
	client      *mollieclient.Client
	webhookURL  string
	redirectURL string
	httpClient  *http.Client
	mu          sync.Mutex
	meta        map[string]paymentMeta
}

type paymentMeta struct {
	ResourcePath string
	Amount       string
	Currency     string
}

func NewProvider(apiKeyEnv, webhookURL, redirectURL string) (*Provider, error) {
	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("mollie api key not set in env %s", apiKeyEnv)
	}
	client := mollieclient.New(mollieclient.WithSecurity(components.Security{
		APIKey: mollieclient.String(apiKey),
	}))
	return &Provider{
		client:      client,
		webhookURL:  webhookURL,
		redirectURL: redirectURL,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
		meta:        make(map[string]paymentMeta),
	}, nil
}

func (p *Provider) CreatePayment(ctx context.Context, req payments.CreatePaymentRequest) (payments.Payment, error) {
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return payments.Payment{}, fmt.Errorf("invalid amount: %w", err)
	}
	description := req.Description
	if description == "" {
		description = fmt.Sprintf("Paywall access: %s", req.ResourcePath)
	}

	paymentReq := &components.PaymentRequest{
		Description: description,
		Amount: components.Amount{
			Value:    fmt.Sprintf("%.2f", amount),
			Currency: req.Currency,
		},
		RedirectURL: mollieclient.String(p.redirectURL),
		WebhookURL:  mollieclient.String(p.webhookURL),
		Metadata: &components.Metadata{
			MapOfAny: map[string]any{"resource_path": req.ResourcePath},
			Type:     components.MetadataTypeMapOfAny,
		},
	}

	var idempotencyKey *string
	if req.IdempotencyKey != "" {
		idempotencyKey = mollieclient.String(req.IdempotencyKey)
	}

	resp, err := p.client.Payments.Create(ctx, nil, idempotencyKey, paymentReq)
	if err != nil {
		return payments.Payment{}, fmt.Errorf("mollie create payment: %w", err)
	}
	payment := resp.GetPaymentResponse()
	if payment == nil {
		return payments.Payment{}, fmt.Errorf("mollie create payment: empty response")
	}

	p.mu.Lock()
	p.meta[payment.ID] = paymentMeta{
		ResourcePath: req.ResourcePath,
		Amount:       req.Amount,
		Currency:     req.Currency,
	}
	p.mu.Unlock()

	return p.toPayment(payment), nil
}

func (p *Provider) GetPayment(ctx context.Context, paymentID string) (payments.Payment, error) {
	resp, err := p.client.Payments.Get(ctx, operations.GetPaymentRequest{PaymentID: paymentID})
	if err != nil {
		return payments.Payment{}, fmt.Errorf("mollie get payment: %w", err)
	}
	payment := resp.GetPaymentResponse()
	if payment == nil {
		return payments.Payment{}, fmt.Errorf("mollie get payment: empty response")
	}
	result := p.toPayment(payment)
	if result.ResourcePath == "" {
		if meta := p.lookupMeta(paymentID); meta.ResourcePath != "" {
			result.ResourcePath = meta.ResourcePath
			result.Amount = meta.Amount
			result.Currency = meta.Currency
		} else if path := metadataResourcePath(payment.Metadata); path != "" {
			result.ResourcePath = path
		}
	}
	if result.Amount == "" && payment.Amount.Value != "" {
		result.Amount = payment.Amount.Value
	}
	if result.Currency == "" && payment.Amount.Currency != "" {
		result.Currency = payment.Amount.Currency
	}
	return result, nil
}

func (p *Provider) CompleteTestPayment(ctx context.Context, paymentID string) error {
	resp, err := p.client.Payments.Get(ctx, operations.GetPaymentRequest{PaymentID: paymentID})
	if err != nil {
		return fmt.Errorf("mollie get payment for test completion: %w", err)
	}
	payment := resp.GetPaymentResponse()
	if payment == nil {
		return fmt.Errorf("payment not found: %s", paymentID)
	}
	if payment.Status == components.PaymentResponseStatusPaid {
		return nil
	}
	links := payment.GetLinks()
	if links.ChangePaymentState == nil || links.ChangePaymentState.Href == nil || *links.ChangePaymentState.Href == "" {
		return fmt.Errorf("changePaymentState link unavailable for %s (test mode only)", paymentID)
	}
	linkHref := *links.ChangePaymentState.Href
	form := url.Values{"status": {"paid"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, linkHref, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("change payment state: %w", err)
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return fmt.Errorf("change payment state returned %d: %s", httpResp.StatusCode, string(body))
	}
	return nil
}

func (p *Provider) toPayment(payment *components.PaymentResponse) payments.Payment {
	checkoutURL := ""
	links := payment.GetLinks()
	if links.Checkout != nil {
		checkoutURL = links.Checkout.Href
	}
	return payments.Payment{
		ID:           payment.ID,
		Status:       mapMollieStatus(payment.Status),
		ResourcePath: metadataResourcePath(payment.Metadata),
		Amount:       payment.Amount.Value,
		Currency:     payment.Amount.Currency,
		CheckoutURL:  checkoutURL,
	}
}

func (p *Provider) lookupMeta(id string) paymentMeta {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.meta[id]
}

func mapMollieStatus(status components.PaymentResponseStatus) payments.Status {
	switch status {
	case components.PaymentResponseStatusPaid:
		return payments.StatusPaid
	case components.PaymentResponseStatusFailed, components.PaymentResponseStatusCanceled, components.PaymentResponseStatusExpired:
		return payments.StatusFailed
	default:
		return payments.StatusOpen
	}
}

func metadataResourcePath(md *components.Metadata) string {
	if md == nil {
		return ""
	}
	if md.MapOfAny != nil {
		if v, ok := md.MapOfAny["resource_path"].(string); ok {
			return v
		}
	}
	return ""
}
