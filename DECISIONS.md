# DECISIONS.md - locked architecture decisions

Short log so a fresh Codex session does not re-litigate settled choices.
Format: Decision - rationale - status.

- **EUR only, for now** - EU-first wedge; multi-currency is post-raise. - LOCKED
- **JSONL append-only ledger, not a database** - zero infra cost; sufficient for the demo;
  DB is post-raise. - LOCKED
- **In-region processing, no raw IPs, no full-log ingestion** - this is the differentiator;
  see DATA-RESIDENCY.md. Non-negotiable. - LOCKED
- **Mollie (EU) is the only external service** - no US endpoints, no extra sub-processors. - LOCKED
- **Mock provider for dev/tests; Mollie test mode for the demo** - no real charges. - LOCKED
- **Grants: Ed25519** - asymmetric signing lets an in-region edge verify grants with
  only the public key; private key is loaded/generated from config. - LOCKED
- **Per-request quota on grants** - grant consumption is jti-keyed and atomic, so access
  can be metered per served request instead of only by flat TTL. - LOCKED
- **DEFERRED until funded:** x402/USDC, ACP, aggregation/batching, Web Bot Auth + routing,
  collecting-society integration, production hosting/DB, dashboards. - DEFERRED

When you make a new structural decision, add a line here in the same format.
