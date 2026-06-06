package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/agentic-paywall/agentic-paywall/internal/approval"
	"github.com/agentic-paywall/agentic-paywall/internal/config"
	"github.com/agentic-paywall/agentic-paywall/internal/gateway"
	"github.com/agentic-paywall/agentic-paywall/internal/guardrails"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Wallet struct {
	cfg      *config.Config
	budget   *guardrails.BudgetTracker
	approval *approval.Prompter
	client   *http.Client
}

func NewWallet(cfg *config.Config) *Wallet {
	return &Wallet{
		cfg:      cfg,
		budget:   guardrails.NewBudgetTracker(cfg.Wallet.DailyBudget),
		approval: approval.NewPrompter(),
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func RegisterTools(server *mcpsdk.Server, wallet *Wallet) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "get_wallet_allowance",
		Description: "Check remaining daily EUR budget before attempting a paywall payment",
	}, wallet.getWalletAllowance)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "execute_paywall_payment",
		Description: "Parse a 402 challenge and initiate payment via the gateway",
	}, wallet.executePaywallPayment)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "verify_transaction_status",
		Description: "Poll grant readiness for a payment and return the access grant token",
	}, wallet.verifyTransactionStatus)
}

type allowanceResult struct {
	DailyBudgetEUR   float64 `json:"daily_budget_eur"`
	RemainingEUR     float64 `json:"remaining_eur"`
	AutoPayThreshold float64 `json:"auto_pay_threshold_eur"`
	Currency         string  `json:"currency"`
}

func (w *Wallet) getWalletAllowance(ctx context.Context, req *mcpsdk.CallToolRequest, _ struct{}) (*mcpsdk.CallToolResult, any, error) {
	result := allowanceResult{
		DailyBudgetEUR:   w.cfg.Wallet.DailyBudget,
		RemainingEUR:     w.budget.Remaining(),
		AutoPayThreshold: w.cfg.Wallet.AutoPayThreshold,
		Currency:         w.cfg.Currency,
	}
	return textResult(result)
}

type executePaymentArgs struct {
	ChallengeJSON string  `json:"challenge_json" jsonschema:"AgentPaywall v1 JSON from 402 body"`
	Amount        string  `json:"amount,omitempty" jsonschema:"EUR amount override"`
	TargetURL     string  `json:"target_url,omitempty" jsonschema:"Original resource URL"`
	Purpose       string  `json:"purpose,omitempty" jsonschema:"Human-readable payment purpose"`
	AmountFloat   float64 `json:"amount_float,omitempty"`
}

type executePaymentResult struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

func (w *Wallet) executePaywallPayment(ctx context.Context, req *mcpsdk.CallToolRequest, args executePaymentArgs) (*mcpsdk.CallToolResult, any, error) {
	challenge, amount, resourcePath, err := w.parseChallenge(args)
	if err != nil {
		return errorResult(err)
	}

	if err := w.budget.CanSpend(amount); err != nil {
		return errorResult(err)
	}
	purpose := args.Purpose
	if purpose == "" {
		purpose = challenge.Resource.Description
	}
	if err := w.approval.RequireApproval(amount, w.cfg.Wallet.AutoPayThreshold, purpose); err != nil {
		return errorResult(err)
	}

	initiateBody := map[string]string{
		"resource_path": resourcePath,
		"amount":        fmt.Sprintf("%.2f", amount),
		"currency":      w.cfg.Currency,
		"description":   purpose,
	}
	body, _ := json.Marshal(initiateBody)
	url := w.cfg.Gateway.BaseURL + "/pay/initiate"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return errorResult(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(httpReq)
	if err != nil {
		return errorResult(fmt.Errorf("gateway pay/initiate: %w", err))
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errorResult(fmt.Errorf("gateway returned %d: %s", resp.StatusCode, string(respBody)))
	}

	var initiateResp struct {
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &initiateResp); err != nil {
		return errorResult(err)
	}

	w.budget.RecordSpend(amount)
	log.Printf("payment initiated: id=%s amount=%.2f EUR resource=%s", initiateResp.PaymentID, amount, resourcePath)

	result := executePaymentResult{
		PaymentID: initiateResp.PaymentID,
		Status:    initiateResp.Status,
		Message:   "Payment initiated. Poll verify_transaction_status for access grant.",
	}
	return textResult(result)
}

type verifyArgs struct {
	PaymentID string `json:"payment_id" jsonschema:"Payment ID from execute_paywall_payment"`
}

type verifyResult struct {
	Ready       bool   `json:"ready"`
	AccessGrant string `json:"access_grant,omitempty"`
	RetryHeader string `json:"retry_header,omitempty"`
}

func (w *Wallet) verifyTransactionStatus(ctx context.Context, req *mcpsdk.CallToolRequest, args verifyArgs) (*mcpsdk.CallToolResult, any, error) {
	if args.PaymentID == "" {
		return errorResult(fmt.Errorf("payment_id required"))
	}
	url := fmt.Sprintf("%s/grants/verify?payment_id=%s", w.cfg.Gateway.BaseURL, args.PaymentID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errorResult(err)
	}
	resp, err := w.client.Do(httpReq)
	if err != nil {
		return errorResult(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return errorResult(fmt.Errorf("gateway returned %d: %s", resp.StatusCode, string(body)))
	}
	var verify gateway.GrantVerifyResponse
	if err := json.Unmarshal(body, &verify); err != nil {
		return errorResult(err)
	}
	result := verifyResult{
		Ready:       verify.Ready,
		AccessGrant: verify.AccessGrant,
		RetryHeader: "PAYMENT-GRANT",
	}
	return textResult(result)
}

func (w *Wallet) parseChallenge(args executePaymentArgs) (gateway.Challenge, float64, string, error) {
	if args.ChallengeJSON == "" {
		return gateway.Challenge{}, 0, "", fmt.Errorf("challenge_json required")
	}
	challenge, err := gateway.ParseChallenge([]byte(args.ChallengeJSON))
	if err != nil {
		return gateway.Challenge{}, 0, "", fmt.Errorf("parse challenge: %w", err)
	}
	if len(challenge.Accepts) == 0 {
		return gateway.Challenge{}, 0, "", fmt.Errorf("challenge has no accepts")
	}
	amountStr := args.Amount
	if amountStr == "" {
		amountStr = challenge.Accepts[0].Amount
	}
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		if args.AmountFloat > 0 {
			amount = args.AmountFloat
		} else {
			return gateway.Challenge{}, 0, "", fmt.Errorf("invalid amount: %w", err)
		}
	}
	resourcePath := challenge.Resource.Path
	if resourcePath == "" && args.TargetURL != "" {
		resourcePath = args.TargetURL
	}
	return challenge, amount, resourcePath, nil
}

func textResult(v any) (*mcpsdk.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{Text: string(data)},
		},
	}, v, nil
}

func errorResult(err error) (*mcpsdk.CallToolResult, any, error) {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{Text: err.Error()},
		},
		IsError: true,
	}, nil, nil
}
