# AGENTS.md — Agentic Content Paywall

Codex reads this before any work. Keep it short; it is loaded on every run.

## What this is
A Go HTTP content gateway that blocks machine clients behind a machine-readable 402
paywall until payment is verified, plus an agent-side wallet MCP server. The strategic
wedge is **EU data residency**: process in-region, never export reader IPs or full logs.
Module path: github.com/agentic-paywall/agentic-paywall

## Layout
- cmd/gateway/main.go        — gateway binary; wires routes, provider, webhook OnPaid
- cmd/payment-mcp/main.go    — agent wallet MCP (stdio) for Cursor/agents
- internal/gateway/          — enforcement core (challenge, middleware, grants, server, handlers, webhook, pay_initiate)
- internal/payments/         — provider.go interface; mollie/, mock/, setup/
- internal/guardrails/budget.go — buyer-side daily EUR spend cap (wallet)
- internal/approval/approval.go — buyer-side Y/N approval over threshold
- internal/mcp/tools.go      — wallet tools
- internal/config/config.go  — JSON config + env overrides

## Build & run (run these YOURSELF; do not make Codex iterate to green)
go build ./...
go vet ./...
go test ./...
./bin/gateway   # provider=mock by default; PAYMENT_PROVIDER=mollie + MOLLIE_API_KEY for test-mode

## Conventions
- Standard library FIRST. Existing deps only: golang-jwt/jwt/v5, google/uuid,
  modelcontextprotocol/go-sdk, mollie/mollie-api-golang. See "Hard rules" before adding any dep.
- Money is decimal strings ("0.50"), currency EUR. NEVER float for money in storage or transport.
- The gateway is the source of truth; the MCP wallet only calls gateway HTTP endpoints.
- Tests: table-driven, standard library `testing` only. See TESTING.md.
- Record locked architecture choices in DECISIONS.md; do not re-litigate them.

## HARD RULES — violating these costs re-runs; do not break them
1. Content is NEVER served without a verified grant. The 402 path must not leak the body.
2. NEVER issue an access grant from an unverified webhook payload. Confirm with
   provider.GetPayment(id) == StatusPaid before issuing.
3. NEVER store raw IP addresses or ingest full HTTP access logs. If an IP is unavoidably
   present, hash it (SHA-256 + config salt) at source. Default: do not collect it.
4. NO new external services or network calls except the EU PSP (Mollie). No US endpoints.
5. Do NOT add a third-party dependency without saying why in the PR description and checking
   it is not already solvable with the standard library. Prefer stdlib.
6. Do NOT use localStorage/sessionStorage or any browser storage (no relevance here, but never).
7. Do NOT log secrets, API keys, or full grant tokens.
8. ONE scoped change per task. No drive-by refactors. Leave build, vet, and test green.

## Scope discipline (budget)
This project runs on a tiny budget. Build ONLY what the active sprint prompt asks for.
Deferred (do not build unless asked): x402/USDC, ACP, aggregation/batching, Web Bot Auth,
production DB, dashboards, hosting. See DECISIONS.md.
