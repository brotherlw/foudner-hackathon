# Sprint 2 — EU data-residency differentiator (T2 reframed)
# Run: codex exec --sandbox workspace-write "$(cat .github/codex/prompts/s2-residency.md)"

Add three things to support an EU data-residency story. Read DATA-RESIDENCY.md first — the
/.well-known endpoint must match its fields.

1) internal/ledger: append-only Event{TS,Type,PaymentID,ResourcePath,Amount,Currency,AgentID,Decision}
   with Ledger interface Append(Event) error; FileLedger (JSONL, path from config, mutex-guarded) and
   MemoryLedger for tests. Emit events from PaywallMiddleware (challenge_issued, access_granted,
   access_denied), PayInitiateHandler (payment_initiated), and the webhook paid path (payment_paid,
   grant_issued). NEVER store raw IP addresses; if an IP is ever present, hash it SHA-256 with a config
   salt; default is to omit it entirely.

2) config: add data_residency { region string, allowed_sub_processors []string, ip_hash_salt string }.
   On startup, fail fast with a clear error if any configured external endpoint host is not in
   allowed_sub_processors.

3) A handler GET /.well-known/data-residency returning JSON {region, collected_fields,
   storage_location, sub_processors, raw_ip_retained:false, cross_border_transfer:false}, plus a func
   that writes the same content to DATA-RESIDENCY-LIVE.md (do not overwrite the human-authored
   DATA-RESIDENCY.md).

Add tests for the ledger (incl. a test asserting no raw IP is ever written) and the residency
guardrail. Follow AGENTS.md and TESTING.md. Money stays decimal strings. Keep build, vet, and test green.
