# REVIEW.md — self-review checklist

Codex: before declaring a task done, review your own diff against every item.
The human will also paste: "Review your diff against REVIEW.md."

## Correctness & security
- [ ] Content is never served without a verified grant; 402 path leaks no body.
- [ ] No grant is issued without provider.GetPayment(id) == StatusPaid.
- [ ] No raw IP addresses stored anywhere; any IP is hashed at source or omitted.
- [ ] No external calls except Mollie (EU). No US endpoints introduced.
- [ ] No secrets, API keys, or full grant tokens logged.

## Scope & cost
- [ ] Exactly one scoped change (no drive-by refactors).
- [ ] No new third-party dependency (or: justified in the PR description, stdlib insufficient).
- [ ] Money stays decimal strings; no floats in storage/transport.

## Quality
- [ ] Table-driven tests added per TESTING.md (happy + failure + idempotency).
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` all green.
- [ ] DECISIONS.md updated if a structural choice was made.

If any box is unchecked, fix it before finishing — do not ask the human to catch it.
