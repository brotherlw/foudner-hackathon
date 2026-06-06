# Sprint 4 — Credibility hardening: Ed25519 grants + per-request quota (T3 + T6)
# NICE-TO-HAVE, only if budget remains. This is the one task worth higher reasoning effort.
# Run: codex exec --sandbox workspace-write "$(cat .github/codex/prompts/s4-harden.md)"

(1) Refactor internal/gateway/grants.go to sign access grants with Ed25519 instead of HS256.
The gateway loads (or generates and persists) an Ed25519 private key via config; serve the public
key at GET /.well-known/agent-paywall-key (base64). Add a package-level
VerifyGrantWithPublicKey(pub ed25519.PublicKey, tokenB64, requestPath string) so an external edge
enforcer can verify a grant with no shared secret. Add a JTI replay store (in-memory, TTL = grant TTL)
that VerifyGrant checks and rejects on reuse.

(2) Add a quota claim to grants and a jti-keyed quota store; PaywallMiddleware decrements per served
request and returns the 402 challenge once quota is exhausted; decrement must be atomic under
concurrent requests.

Add tests for both (valid, tampered, path-mismatch, expired, replayed-jti, quota-exhausted,
concurrent-no-overserve). Update DECISIONS.md to mark these LOCKED. Follow AGENTS.md and TESTING.md.
Keep build, vet, and test green. Do NOT use best-of-N.
