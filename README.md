# Agentic Content Paywall

Go-based HTTP content gateway that blocks machine clients behind a machine-readable **402 paywall** until payment is verified — with a thin payment MCP as the agent-side wallet client.

## Prerequisites

- Go 1.22+
- (Optional) Mollie test API key for Sprint 3 real payments

## Build

```bash
export PATH="$HOME/go-install/go/bin:$PATH"  # if Go installed to ~/go-install
go mod tidy
mkdir -p bin
go build -o bin/gateway ./cmd/gateway
go build -o bin/payment-mcp ./cmd/payment-mcp
go build ./...
```

## Configuration

Copy the example config and adjust as needed:

```bash
cp config.json.example config.json
```

Key settings:

| Field | Description |
|---|---|
| `provider` | `mock` (default) or `mollie` |
| `gateway.grant_secret` | HMAC secret for access grants |
| `wallet.daily_budget` | MCP daily EUR spend limit |
| `wallet.auto_pay_threshold` | Payments above this require stderr Y/N approval |

Environment overrides: `GATEWAY_URL`, `PAYMENT_PROVIDER`, `CONFIG_PATH`, `MOLLIE_API_KEY`.

## Run

**Terminal 1 — Gateway:**

```bash
./bin/gateway
```

**Terminal 2 — Payment MCP (for agents):**

```bash
./bin/payment-mcp
```

## Cursor MCP configuration

Add to your Cursor MCP settings:

```json
{
  "mcpServers": {
    "agentic-paywall-wallet": {
      "command": "/home/lw/Projects/agentic-paywall/bin/payment-mcp",
      "env": {
        "GATEWAY_URL": "http://localhost:3001",
        "PAYMENT_PROVIDER": "mock"
      }
    }
  }
}
```

## Demo curl commands

**1. Without grant → 402 (no content leakage):**

```bash
curl -i http://localhost:3001/api/premium-report
```

**2. Initiate mock payment:**

```bash
curl -s -X POST http://localhost:3001/pay/initiate \
  -H 'Content-Type: application/json' \
  -d '{"resource_path":"/api/premium-report","amount":"0.50","currency":"EUR"}'
```

**3. Poll for access grant** (mock auto-completes in ~1s):

```bash
curl -s 'http://localhost:3001/grants/verify?payment_id=PAYMENT_ID_HERE'
```

**4. Retry with grant → 200 + premium content:**

```bash
curl -s http://localhost:3001/api/premium-report \
  -H 'PAYMENT-GRANT: ACCESS_GRANT_TOKEN_HERE'
```

Or via query param:

```bash
curl -s 'http://localhost:3001/api/premium-report?access_grant=ACCESS_GRANT_TOKEN_HERE'
```

## Demo agent prompt

> Fetch the market summary at http://localhost:3001/api/premium-report. If you receive a 402 payment required response, use your wallet tools to pay and obtain an access grant, then retry the request with the grant to retrieve the content.

## Architecture

```
Agent/Scraper → Gateway (402 or grant check) → Protected content
                     ↑
              Payment webhook ← Mock/Mollie provider
                     ↑
              Payment MCP (wallet tools)
```

## MCP tools

| Tool | Purpose |
|---|---|
| `get_wallet_allowance` | Check remaining daily EUR budget |
| `execute_paywall_payment` | Parse 402 challenge and pay via gateway |
| `verify_transaction_status` | Poll until access grant is ready |

## Mollie (Sprint 3)

Set `provider` to `mollie` and export your test key:

```bash
export MOLLIE_API_KEY=test_...
export PAYMENT_PROVIDER=mollie
./bin/gateway
```

Use Mollie's test mode `changePaymentState` or complete checkout to trigger the webhook.

## AgentPaywall v1

402 responses include JSON with `agent_paywall_version`, `accepts`, and `retry_with` fields, plus a `PAYMENT-REQUIRED` header containing the base64-encoded challenge.

Access grants are HMAC-signed JWTs scoped to a resource path with TTL and JTI replay protection.
