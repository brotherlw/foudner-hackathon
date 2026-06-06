# Sprint Build Plan — Codex on a $100 budget

Build plan for `github.com/brotherlw/foudner-hackathon` (module `github.com/agentic-paywall/agentic-paywall`), executed with **OpenAI Codex**, in sprints, on a **$100 total budget**.

> **Read this first.** $100 buys a **fundable demo, not a production launch.** The single goal of every sprint below is one outcome: *an agent pays for EU publisher content end-to-end, in EUR, with the whole flow processed and logged in-region — no reader IPs, nothing leaving the EU.* That demo is your wedge against TollBit (US, exports reader IPs) and your pitch to a publisher's DPO. Everything that doesn't serve that demo is **deferred until funded.**

---

## 0. Budget reality & the rules that keep you under $100

**Where the money goes (keep infra at $0):**

| Item | Cost | Note |
|---|---|---|
| Codex access | ~$40–60 | ChatGPT Plus includes Codex (~$20/mo) — 2–3 months. Or load equivalent as API credit. This is your only real spend. |
| Infra (dev + demo) | **$0** | Local dev; Mollie **test mode** (free); free-tier host for the live demo (Fly.io / Render free, or ngrok over localhost). EU region only. |
| Domain (optional) | ~$10 | A `.eu` domain makes the residency story tangible. Skip if tight. |
| Buffer | ~$30 | Re-runs, overage. |

**Codex frugality rules (this is how $100 lasts — every rule saves tokens):**
1. **Commit `AGENTS.md` first** (see §1) so Codex never re-explores the repo. Re-exploration is the biggest hidden cost.
2. **One task = one tightly-scoped prompt = one PR.** No open-ended "improve the codebase" prompts — they burn tokens looping.
3. **Run `go build` / `go test` / `go vet` YOURSELF, locally (free).** Don't make Codex iterate to green — paste failures back only if stuck.
4. **Use `codex exec` with prompt files** (reproducible, no chatty back-and-forth). Store them in `.github/codex/prompts/`.
5. **Medium reasoning for boilerplate; high reasoning only for the 1–2 genuinely hard tasks.** Don't pay for deep reasoning on a config struct.
6. **Never use `--attempts`/best-of-N** except once, on the hardest task. It multiplies cost N×.
7. **Write tests once;** let them catch regressions instead of re-prompting Codex to re-check.
8. **Keep files small.** Big generated files = big diffs = big token bills on the next edit.

---

## 1. Sprint 0 — Foundation (mostly you, ~$0 Codex)

**Objective:** make the repo Codex-ready and the loop runnable locally, so later sprints don't waste tokens on setup.

**Do yourself (no Codex):**
- Commit the `AGENTS.md` from the earlier Codex build-instructions doc to repo root.
- `cp config.json.example config.json`; `go mod tidy && go build ./...` green.
- Get a **Mollie test API key** (free); confirm `provider=mock` runs, then `PAYMENT_PROVIDER=mollie` with the test key.
- Create `.github/codex/prompts/` for the prompt files below.
- Install Codex: `npm i -g @openai/codex`; `codex` to auth.

**Exit check:** the four demo curl commands in the README work against the mock provider. **Spend: ~$0.**

---

## 2. Sprint 1 — Make it correct & safe (MUST · ~$ low)

**Objective:** fix the one thing that makes the demo a liability — the webhook trusts its payload — and establish a test baseline. (This is task **T1**.)

**Why it's first:** if a forged webhook mints a free grant, your "we built secure payment infrastructure" pitch dies in the first technical question. Cheap to fix, high credibility.

**In scope:** webhook payment verification + `internal/gateway/webhook_test.go`.
**Out of scope:** everything else.

**Codex prompt** (`prompts/s1-webhook-verify.md`):
```text
Read internal/gateway/webhook.go, internal/gateway/server.go, internal/payments/provider.go,
cmd/gateway/main.go. Security bug: the webhook issues an access grant without confirming the
payment was actually paid. Fix: in the OnPaid path, always call provider.GetPayment(payload.PaymentID)
and only issue a grant when status == StatusPaid. Treat unknown IDs and non-paid statuses as a
no-op returning HTTP 200 {"ok":true,"ignored":true} (don't reveal whether the ID exists). Make
grant issuance idempotent per payment_id. Add internal/gateway/webhook_test.go (table-driven, mock
provider): forged-unpaid → no grant, genuine-paid → grant, duplicate → one grant. Standard library
testing only. Keep go build ./... and go test ./... green.
```
**You run locally after:** `go test ./... && go vet ./...`.
**Exit check:** forged webhook → no access. **Effort: ~1 short Codex run.**

---

## 3. Sprint 2 — The data-residency differentiator (MUST · the wedge)

**Objective:** build the thing no US competitor has — verifiable in-region processing with no reader-IP exfiltration — and make it *demonstrable*. (This is **T2** reframed around residency.)

**Why it matters most:** this is your entire pitch. TollBit forces publishers to export all HTTP logs incl. IPs to the US. You collect almost nothing, store it in-region, and can prove it. Build the proof.

**In scope (three small pieces):**
1. **In-region append-only ledger** — `internal/ledger/` writing JSONL to a configurable EU path. Event = `{ts, type, payment_id, resource_path, amount, currency, agent_id, decision}`. **No raw IPs** — if any IP is ever handled, hash it (SHA-256 + salt) at source; default is to not store it at all.
2. **Residency config + guardrail** — `config.data_residency.region` (e.g. `eu`) and a startup check that refuses to boot if a non-EU sub-processor/endpoint is configured. The only external call allowed is the EU PSP (Mollie).
3. **The attestation artifact** — generate a machine-readable `data-flow manifest` (served at `/.well-known/data-residency` and written to `DATA-RESIDENCY.md`) listing: what's collected, where it's stored, sub-processors (Mollie, EU), and an explicit "no cross-border transfer / no raw IP retention" statement. **This is your DPO sales collateral, generated from config.**

**Out of scope:** real database (JSONL is enough for the demo), aggregation/batching, dashboards.

**Codex prompt** (`prompts/s2-residency.md`):
```text
Add three things to support an EU data-residency story.
1) internal/ledger: append-only Event{TS,Type,PaymentID,ResourcePath,Amount,Currency,AgentID,Decision}
   with Ledger interface Append(Event) error; FileLedger (JSONL, path from config, mutex) + MemoryLedger
   for tests. Emit events from PaywallMiddleware (challenge_issued, access_granted/denied),
   PayInitiateHandler (payment_initiated), webhook paid path (payment_paid, grant_issued).
   NEVER store raw IP addresses; if an IP is ever present, hash it SHA-256 with a config salt. Default: omit it.
2) config: add data_residency { region string, allowed_sub_processors []string }. On startup, fail fast
   with a clear error if any configured external endpoint host is not in allowed_sub_processors.
3) A handler at GET /.well-known/data-residency returning JSON {region, collected_fields, storage_location,
   sub_processors, raw_ip_retained:false, cross_border_transfer:false}, and a command/func that writes the
   same content to DATA-RESIDENCY.md. Add tests for the ledger and the residency guardrail. Money stays
   decimal strings. Keep build and tests green.
```
**Run locally:** `go test ./...`; hit `/.well-known/data-residency`; check `cat DATA-RESIDENCY.md`.
**Exit check:** ledger has events, contains **no raw IPs**, and the residency endpoint renders. **Effort: 1 medium Codex run** (use medium reasoning).

---

## 4. Sprint 3 — The end-to-end demo loop (MUST · what you show)

**Objective:** make the full agent→402→pay(EUR via Mollie test)→grant→serve loop run, logging to the in-region ledger. This is the recording you put in the deck.

**In scope:** wire the **real Mollie EUR test-mode** path through the gate (not just mock); confirm the MCP wallet drives it; write a `DEMO.md` runbook + a `make demo` / shell script that runs the whole loop.
**Out of scope:** UI, hosting beyond a free tier / ngrok, x402.

**Codex prompt** (`prompts/s3-demo.md`):
```text
Make the end-to-end loop runnable and reproducible with the Mollie test provider in EUR.
Verify config.provider=mollie path: agent requests a protected resource → 402 with AgentPaywall
challenge → /pay/initiate creates a Mollie test payment → CompleteTestPayment (changePaymentState) →
webhook verifies paid (Sprint 1) → grant issued → retry with grant → 200 + content, with every step
appended to the in-region ledger (Sprint 2). Fix any wiring gaps. Add a scripts/demo.sh that runs the
full loop against a local gateway and prints each step, and a DEMO.md runbook. Do not add paid
dependencies. Keep build and tests green.
```
**Run locally:** `./scripts/demo.sh` end to end; then **record it** (terminal capture).
**Exit check:** one command shows pay→access in EUR, ledger updated, no IPs. **Effort: 1 Codex run + your recording.**

> After Sprint 3 you have a complete, demoable, defensible product slice. **If money/time runs out here, you can still pitch.** Sprints 4+ are credibility upgrades, not blockers.

---

## 5. Sprint 4 — Credibility hardening (NICE-TO-HAVE · only if budget left)

**Objective:** the two upgrades that make a technical investor nod. (Tasks **T3** + **T6**.)

- **Ed25519 grants + JTI replay store** — asymmetric signing so an in-region edge can verify without a shared secret (supports the "in-region edge enforcement" story); reject replayed JTIs.
- **Per-request quota** — grant carries a decrementing quota so "metering" is real, not a flat window.

**Codex prompt** (`prompts/s4-harden.md`):
```text
(1) Refactor internal/gateway/grants.go to sign grants with Ed25519 (load/generate+persist key via
config), serve the public key at GET /.well-known/agent-paywall-key, add VerifyGrantWithPublicKey for
an external edge, and add a JTI replay store (in-memory, TTL=grant TTL) that VerifyGrant checks.
(2) Add a quota claim to grants and a jti-keyed quota store; PaywallMiddleware decrements per serve,
returns 402 when exhausted, atomic under concurrency. Add tests for both. Keep build and tests green.
```
**This is the one task worth higher reasoning effort.** Do NOT use best-of-N. **Effort: 1 careful Codex run.**

---

## 6. Deferred until funded (do NOT build now)

These are real roadmap items but they don't change the $100 demo, and each is a token/time sink:
- x402 / USDC agent lane and ACP Delegate-Payment.
- Seller-side aggregation / batching (needs verified agent identity first).
- Web Bot Auth + the full detection/routing trichotomy.
- Collecting-society integration (the "enforcement layer under collective licensing" channel — high-value post-raise).
- Production DB, multi-tenant dashboard, real hosting, SOC2/PCI work.

Put these on the deck as "post-raise roadmap," not in the build.

---

## 7. Definition of done (every sprint)

- `go build ./...`, `go vet ./...`, `go test ./...` green **(you run these locally, free).**
- New behavior has table-driven tests.
- No raw IPs stored anywhere; nothing leaves the EU except the Mollie (EU) call.
- One PR per sprint; skim the diff before merge (`git diff`), `/review` only if something feels off.

## 8. The demo you're building toward (keep this taped to your monitor)

> A coding agent (via the wallet MCP) requests a gated EU publisher article → gets a 402 in EUR → pays through Mollie (EU) → receives an in-region access grant → reads the article. The ledger shows the transaction; `/.well-known/data-residency` proves no reader IPs were stored and nothing crossed the EU border. *"TollBit makes the publisher export reader IPs to the US to do this. We don't."*

## 9. Frugality checklist (tick before each Codex run)

- [ ] `AGENTS.md` committed (Codex won't re-explore)
- [ ] Prompt is one scoped task with explicit acceptance + tests
- [ ] I'll run build/test locally, not via Codex
- [ ] Medium reasoning unless this is Sprint 4
- [ ] No `--attempts` / best-of-N
- [ ] Prompt saved to `.github/codex/prompts/` for reuse

---

*Pairs with the task specs in the earlier Codex build-instructions doc (T1–T6 map to Sprints 1–4). Verify Codex flags against `codex --help` and current ChatGPT plan inclusions, as pricing and CLI options change.*
