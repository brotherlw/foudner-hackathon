# TESTING.md — test conventions

Codex: follow these so tests stay cheap, consistent, and dependency-free.

## Rules
- Standard library `testing` only. Do NOT add testify, ginkgo, or any test framework.
- Table-driven tests. One `t.Run(tc.name, ...)` per case.
- Name tests `TestThing_case` (e.g. `TestWebhook_forgedUnpaid_noGrant`).
- Use the mock payment provider (internal/payments/mock) for anything touching payments —
  never call the real Mollie API in a test.
- For HTTP handlers, use net/http/httptest. For external directories/keys, spin a local
  httptest.Server as a stub; never hit the network.
- Each PR must leave `go test ./...` and `go vet ./...` green. The author (human) runs them.

## What to cover (minimum per task)
- The happy path.
- The security-relevant failure path (e.g. unpaid/forged → no grant; no token → 402; tampered → reject).
- One idempotency / replay case where state changes (e.g. duplicate webhook → one grant).

## What NOT to do
- No flaky time.Sleep-based tests; inject a clock or use deterministic values.
- No tests that depend on wall-clock timezone or on files outside the test's temp dir.
- No raw IPs in fixtures.
