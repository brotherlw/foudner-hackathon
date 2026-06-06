# Sprint 1 — Verify payments before issuing grants (T1, security)
# Run: codex exec --sandbox workspace-write "$(cat .github/codex/prompts/s1-webhook-verify.md)"

Read internal/gateway/webhook.go, internal/gateway/server.go, internal/payments/provider.go,
and cmd/gateway/main.go. Security bug: the webhook issues an access grant without confirming the
payment was actually paid. Fix: in the OnPaid path, always call provider.GetPayment(payload.PaymentID)
and only issue a grant when status == StatusPaid. Treat unknown IDs and non-paid statuses as a no-op
returning HTTP 200 {"ok":true,"ignored":true} (do not reveal whether the ID exists). Make grant
issuance idempotent per payment_id. Add internal/gateway/webhook_test.go (table-driven, mock provider,
stdlib testing only): forged-unpaid -> no grant; genuine-paid -> grant present; duplicate webhook ->
exactly one grant. Follow AGENTS.md hard rules and TESTING.md. Keep go build ./... and go test ./... green.
