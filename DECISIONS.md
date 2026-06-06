# DECISIONS.md — locked architecture decisions

Short log so a fresh Codex session does not re-litigate settled choices.
Format: Decision — rationale — status.

- **EUR only, for now** — EU-first wedge; multi-currency is post-raise. — LOCKED
- **JSONL append-only ledger, not a database** — zero infra cost; sufficient for the demo;
  DB is post-raise. — LOCKED
- **In-region processing, no raw IPs, no full-log ingestion** — this is the differentiator;
  see DATA-RESIDENCY.md. Non-negotiable. — LOCKED
- **Mollie (EU) is the only external service** — no US endpoints, no extra sub-processors. — LOCKED
- **Mock provider for dev/tests; Mollie test mode for the demo** — no real charges. — LOCKED
- **Grants: HS256 today; Ed25519 deferred to Sprint 4** — asymmetric signing enables
  edge verification but is not needed for the core demo. — PLANNED (Sprint 4)
- **Per-request quota deferred to Sprint 4** — flat-window grant is fine for the demo. — PLANNED (Sprint 4)
- **DEFERRED until funded:** x402/USDC, ACP, aggregation/batching, Web Bot Auth + routing,
  collecting-society integration, production hosting/DB, dashboards. — DEFERRED

When you make a new structural decision, add a line here in the same format.
