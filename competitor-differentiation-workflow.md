# Content-Access Payments — Competitor & Differentiation Workflow

*A staged map of who already exists at each layer of the "agents/users pay to access content" stack, and the concrete differentiation move at each stage.*

> **How to read this:** The product splits into 7 sequential layers. At each layer you either **adopt** an existing standard/competitor, or you **differentiate**. The workflow flags which is which. Defensibility lives in Layers 3 and 5–7, not Layers 1–2.

---

## STAGE 0 — Frame the battlefield

**The transaction lifecycle you operate in:** `Discover → Authorize → Block/Gate → Pay → Settle → Reconcile → Pay out`

**Strategic warning (decide before building):**
Durable value is expected to accrue to *settlement rails and facilitators*, not to the open protocols themselves (the TCP/IP-captured-no-rents argument). A thin wrapper over someone else's protocol is exposed. **→ Own the publisher relationship + the enforcement gate + the reconciliation ledger, not "another rail."**

```
                 ┌──────────────────────────────────────────────────────┐
   AGENT / USER  │  L1 Connect → L2 License → L3 Block/Gate → L4 Pay →   │  PUBLISHER
      ──────────▶│  L5 Aggregate → L6 Reconcile → L7 Pay out             │──────────▶
                 └──────────────────────────────────────────────────────┘
                    ADOPT (commodity)              DIFFERENTIATE (moat)
```

---

## STAGE 1 — Connection layer (how the agent reaches you)

**Decision:** Adopt, do not build.

| Competitor / Standard | What it is | Status |
|---|---|---|
| **MCP (Anthropic)** | De-facto agent tool/data connection standard | ~97M monthly SDK downloads, 10k+ public servers |
| **UCP (Google)** | Commerce-journey standard (discovery→checkout) | Endorsed by 20+ partners (Shopify, Adyen, Stripe, Visa…) |
| **ACP (OpenAI + Stripe)** | Delegated-payment checkout standard | Stripe built reference impl; works with any delegated-token PSP |
| **A2A / AP2 (Google)** | Agent-to-agent + agent payment authorization | Complementary to UCP/MCP |

**Workflow action:**
- ✅ Expose an **MCP server** so any agent can find and call your paywall.
- ✅ Stay compatible with UCP/ACP rather than competing.
- ❌ Do **not** invent a new connection protocol — pure commodity, zero moat.

**Differentiation here:** none. This is table stakes.

---

## STAGE 2 — Licensing layer (terms for access)

**Decision:** Adopt a standard, add a value layer on enforcement + carve-outs.

| Competitor / Standard | Model | Strength | Weakness |
|---|---|---|---|
| **RSL (Really Simple Licensing)** | pay-per-crawl, pay-per-inference, subscription; machine-readable in robots.txt; supports **encrypted/paywalled** content | Broad publisher backing (Reddit, BuzzFeed, USA Today, Vox); CDN support (Cloudflare, Akamai, Fastly); collective licensing | **Voluntary compliance** — no major AI co. committed at launch |
| **Cloudflare Pay-Per-Crawl** | Per-access micro-fee, block-by-default | Technically enforceable at CDN edge | Locks you to Cloudflare CDN |
| **Direct deals (NYT, News Corp)** | Bilateral licensing | High value per deal | Doesn't scale to long-tail publishers |

**Workflow action:**
- ✅ Speak **RSL** for terms (especially its encrypted/paywalled-content support).
- ✅ Solve RSL's enforcement gap — pair declarative terms with **hard payment-gating** (you don't trust voluntary compliance; you block until paid). *(→ Stage 3.)*
- ✅ Build **carve-outs for research/non-profit/academic** access (Creative Commons warns against an "elitist web" — this is both ethical and good positioning).

**Differentiation here:** *enforcement teeth* + *tiered access ethics* on top of an open standard.

---

## STAGE 3 — Content Blocking / Enforcement ⭐ THE GATE (moat-adjacent)

**Decision:** Build. This is where "please pay" becomes "you don't get the bytes." It is the **same pipeline as reconciliation (Stage 6), viewed from the other end** — the gate is the producer, the ledger is the consumer.

**Cardinal rule — three separate problems. Never merge them.**

| # | Problem | Question | Tooling | Spoofable? |
|---|---|---|---|---|
| 1 | **Detection** | Who is knocking? | Web Bot Auth (RFC 9421 sigs) > reverse DNS > UA string > robots.txt | identity layer is crypto-hard; the rest trivially spoofed |
| 2 | **Authorization** | Is this visitor allowed to read this URL? | YOUR policy (RSL terms + price table) | n/a — it's policy |
| 3 | **Enforcement** | The actual gate | 402 → token → quota loop at edge/origin | this is the teeth |

> ⚠️ **Web Bot Auth is identity, NOT authorization.** A verified, signed agent can still scrape paid content if your rules let it through. Don't buy the identity hype and skip the gate — your product lives in Problem 3.

**Detection stack (weakest → strongest):**
- `robots.txt` / `llms.txt` — declarative courtesy sign; ~half of AI traffic ignores it. Never a lock.
- **User-agent string** — trivially spoofed.
- **Reverse DNS / IP allowlist** — legacy; known crawlers only.
- **Web Bot Auth** — Ed25519 + HTTP Message Signatures, public keys at a `.well-known` directory. Backed by Cloudflare, Akamai, Amazon, Google (experimental). Crypto-hard, but **identity-only** and still an IETF draft.

**The trichotomy you MUST encode (not "bot vs human"):**
1. **Background training crawler** (no human) → prime charge/block target.
2. **User-triggered agent fetcher** (e.g. Google-Agent) → gray zone; treated as a browser visit, ignores robots.txt. Your pricing must take a (contestable) position: is "human asks ChatGPT to read this" a human visit or agent access?
3. **Human browser** → existing paywall/subscription logic.

### Routing, not gatekeeping: bot/human classification

> ⚠️ **Inverted threat model.** Mainstream bot detection (CAPTCHA, fingerprinting) exists to *keep bots out*. Here the bot is the **paying customer** — copying the adversarial toolkit walls out your own revenue. The classifier's job is **routing** (which payment lane?), not **gatekeeping** (prove you're human or leave).

**The 402 handshake IS the classifier — passive, no challenge:**
- A bot/agent understands a 402 → answers with a payment token → **machine-pay path**.
- A human browser can't speak "402-pay" → your JS catches the 402 → renders the **human paywall / subscription UI**.
- Web Bot Auth makes honest agents *self-identify* — a signed request *wants* to announce itself, because identifying is how it pays. You don't hunt them; they knock on the toll booth.

> 🔑 **THE ONE RULE:** Never challenge a request that carries a valid payment token OR a valid Web Bot Auth signature. Payment is a stronger "legitimate client" signal than any proof-of-humanity. **Let the money talk.** Any challenge lives only on the human-suspected path, and is invisible/risk-based — never a classic CAPTCHA.

**Cursor movement / behavioral signals — weak, last-resort only:**
- **Browser-only:** pure HTTP / API agents move no cursor; absence ≠ human, presence ≠ human (stealth browsers simulate human curves).
- **Spoofable:** the single most-faked signal — naive bots show constant velocity / straight lines, but stealth tooling fakes non-linear human curves.
- **GDPR:** mouse-movement biometrics = personal data → needs lawful basis + disclosure (real liability for an EU-first product).
- → Use as **one input to a passive risk score** on the ambiguous middle. Never a gate, never a challenge, never in the token-bearing path.

**Routing logic:**
```
valid Web Bot Auth signature?   → KNOWN BOT     → machine pay path (no challenge, ever)
valid payment token?            → PAYING BOT    → serve + meter   (no challenge, ever)
executes JS + browser session   → LIKELY HUMAN  → human paywall / subscription UI
pure HTTP, no token, no sig      → UNVERIFIED    → 402 invite-to-pay, or block per policy
ambiguous headful, no sig/token → RISK-SCORE    → (cursor = 1 input) → invisible challenge
                                                   ONLY if the human lane is free/cheap
```

**The real adversary (inversion of the inversion):** not "a human gets blocked" but "a freeloader bot fakes *humanity* to take the cheaper/free human lane." This only bites where the human lane is free (e.g. an open ad-supported article). **Fix is economic, not technical:** make the verified-bot path so cheap + frictionless (instant sub-cent, no account) that paying beats evading. You out-design the arms race by removing the reason to fight it. Reserve behavioral detection for free-human-lane abuse only, and price in residual leakage rather than CAPTCHA-walling your customers.

**The gate-flow (402 → scoped-token → quota-decrement):**
```
1. Agent requests GET /article         → PEP intercepts BEFORE origin
2. PEP looks up price (RSL/price table) → returns 402 + crawler-price header
3. Agent calls pay API                 → token service mints short-lived SIGNED
                                          token scoped to zone + byte/req quota
4. Agent retries WITH token            → PEP verifies signature + remaining quota
5a. Valid  → fetch origin, stream, decrement quota, EMIT access event ─┐
5b. Invalid → blocked at edge; nothing leaves origin                   │
                                                                       ▼
                                            → reconciliation ledger (Stage 6)
```
Design philosophy: **"freeloaders die at the edge"** — content never leaves the origin until paid. If your gate sits at the origin only, the request has already cost you and the content is one bug away from leaking.

**Headers in play:** `signature-agent` / `signature-input` / `signature` (identity); `crawler-price` (ask) / `crawler-exact-price` (accept) / `crawler-max-price` (budget cap).

**Competitor reality — Cloudflare already owns the edge + merchant-of-record position:**

| Competitor | Model | Your gap to exploit |
|---|---|---|
| **Cloudflare Pay-Per-Crawl / AI Crawl Control** | block-by-default, 402, CF = merchant of record, 1B+ 402s/day | locks publisher to CF CDN; CF settlement, not EU rails |
| **Fastly / Akamai gatekeeper (RSL)** | edge bot validation | CDN-bound |
| **xpay / Zuplo** | x402 paywall-as-a-service | thin wrapper; no EU fiat; no reconciliation |

**Differentiation here:** **edge-OR-origin-flexible enforcement** (not locked to one CDN) + EU-rails settlement + the access-event stream feeding *your own* reconciliation ledger. Don't out-Cloudflare Cloudflare — serve the publishers it can't (non-CF sites) with the rails it doesn't (EU) and the layer it doesn't own (cross-publisher reconciliation/licensing).

**Hard tradeoffs to decide up front:**
- **False positives / SEO:** blocking Googlebot or a paying reader is worse than leaking to a freeloader. The trichotomy is your guardrail.
- **Evasion arms race:** human-mimicking scrapers up ~300% since 2023; residential/mobile proxies defeat IP checks. Crypto identity handles *honest* agents; *dishonest* ones need behavioral defense (expensive — buy, don't build from scratch). Price the residual leakage in.
- **Legal:** block-by-default-then-charge is safer in the EU (Data Act) than silent scraping — favors your EU posture.

---

## STAGE 4 — Payment layer (the actual charge)

**Decision:** Dual-rail. This is where you split from Stripe.

| Competitor | Rail | Fits micropayments? | Geography |
|---|---|---|---|
| **Stripe / ACP / Tempo** | Card + own blockchain | ❌ card fees kill sub-cent; Tempo = their answer | US-centric |
| **x402 (Coinbase) + USDC** | HTTP 402 stablecoin, onchain, ~2s settle | ✅ near-zero fee, sub-cent viable | Global/crypto |
| **xpay / Zuplo** | x402 paywall-as-a-service | ✅ but thin wrapper | Global/crypto |
| **Worldpay MCP / Nexi MCP / Worldline** | PSP MCP servers | ⚠️ fiat fee floor | Worldpay global; **Nexi/Worldline = EU** |
| **Mollie** | SEPA Direct Debit, iDEAL, Bancontact, card, recurring | ❌ raw, ✅ when **aggregated** | **EU-native** |

**The Stripe differentiation (3 wedges):**
1. **Micropayment economics** — cards are uneconomic at €0.001–0.10/article; even Stripe concedes this (it built Tempo). You win with metering + aggregation, not generic checkout.
2. **European-first fiat** — Mollie (SEPA/iDEAL/Bancontact) fits EU publishers + EU regulation; Stripe/ACP are US-shaped.
3. **Content-specific primitives** — per-article unlock, metering, agent-vs-human identity. Stripe gives generic charges only.

**ACP note (the two payment paths):** ACP is OpenAI+Stripe but PSP-agnostic via **Delegate Payment** (vault token, your PSP) vs Stripe-locked via **Shared Payment Token (SPT)**. Direct Delegate-Payment integration needs PSP / PCI-DSS-L1 status. **→ Use ACP's discrete-purchase lane via Delegate Payment on Mollie if/when supported; don't let SPT make you Stripe-dependent.**

**Workflow action:**
- ✅ **Agent / crypto-native flow →** x402 + USDC for true per-request micropayments.
- ✅ **Human / recurring / EU flow →** Mollie mandate (first payment → recurring SEPA/card); ACP Delegate-Payment lane for ChatGPT buyers.
- ❌ Don't charge raw micro-amounts through any card rail.

---

## STAGE 5 — Aggregation layer (make micro-amounts viable) ⭐ MOAT

**Decision:** Build. No incumbent owns this for content.

**The problem:** thousands of €0.002 access events cannot each be a fiat transaction.

**The pattern (matches where x402 itself is heading — spend caps + delayed settlement):**
```
meter many tiny accesses
        │
        ▼
apply spending cap  ("up to €5 for this session/day")
        │
        ▼
batch → ONE settled charge (SEPA mandate / card / USDC sweep)
```

**Workflow action:**
- ✅ Meter each access event (content unit + agent/user + license term). *(Source = Stage 3 gate.)*
- ✅ Enforce per-session/per-day caps. *(Same token-quota the gate already issues.)*
- ✅ Settle periodically as one larger charge.
- ✅ This same design **also satisfies PSD2 SCA** low-value / merchant-initiated exemptions — architecture and compliance are the same decision.

**Differentiation here:** the aggregation engine is the thing that makes fiat micropayments real. Wrappers like xpay don't do this for EU fiat.

---

## STAGE 6 — Reconciliation layer ("transaction recognition") ⭐ CORE MOAT

**Decision:** Build. Hardest part = highest defensibility.

**What it must match:**
`settled payment → which content unit → which publisher → which agent/user → which license term`

**Edge cases that make it hard (and defensible):**
- Refunds
- Failed SEPA debits
- Chargebacks
- Partial / disputed access

**Workflow action:**
- ✅ Consume the **access-event stream from Stage 3** as the source of truth → aggregate → match to settlements.
- ✅ Use Mollie's payment **metadata fields**, **webhooks** (status changes), and **settlements** reporting; reconcile when settlements land.
- ✅ Treat the reconciliation engine as a *feature*, not plumbing.

**Differentiation here:** Stripe-style generic checkout gives you none of this. This is the product.

---

## STAGE 7 — Payout + compliance layer

**Decision:** Offload regulated burden to a licensed PSP; own the reporting.

**Two regimes apply:**

**A) Payments (near-term, bigger burden)**
- PSD2 / **SCA** → handled by the Stage-5 aggregation design + exemptions.
- **PCI-DSS** → minimize scope; let Mollie host card capture.
- **GDPR** → data-minimization on per-access logs (the Stage 3 event stream).
- ✅ Using Mollie as licensed PSP offloads payment-institution burden — a real reason to build on a PSP vs. handling funds directly.

**B) EU AI Act (moving target)**
- **Article 50 transparency** applies from **2 Aug 2026** (possible extension to 2 Dec 2026 for some generative systems).
- Agentic AI must **disclose it's an AI** when interacting with people; if unclear whether it'll meet a human, default to disclosing.
- Annex III high-risk obligations deferred to **Dec 2027** (omnibus, not all formally adopted — treat dates as provisional).
- ✅ If you're the *payment/metering/gate layer* (not the agent), you're likely a downstream component, not a high-risk AI system — **but classify early** if you add agent decision logic.

---

## IMPLEMENTATION STATUS — repo ↔ workflow map

*Repo: `github.com/brotherlw/foudner-hackathon` — "Agentic Content Paywall" (Go). A clean hackathon MVP of the **spine**: the Stage 3 enforcement gate + the Stage 4 Mollie fiat lane, fronted by an agent-side wallet MCP. The moat (5–6) and most differentiation (2, parts of 3–4) layers are not built yet.*

| Stage | Repo artifact | Status |
|---|---|---|
| **L1 Connect** | `cmd/payment-mcp`, `internal/mcp/tools.go` (tools: `get_wallet_allowance` / `execute_paywall_payment` / `verify_transaction_status`) | ✅ Built — but it's the **agent-side wallet** MCP, not a seller-side MCP exposing the paywall to agents. Seller side is plain HTTP. |
| **L2 License** | — | ❌ Missing. Uses a custom **AgentPaywall v1** challenge schema, not RSL. No pay-per-crawl/inference terms, no carve-outs. |
| **L3 Block / Enforce** | `gateway/{middleware,challenge,grants,handlers,server}.go` | ✅ **Core built:** 402 challenge (JSON body + base64 `PAYMENT-REQUIRED` header), grant check, content **dies at origin** (no leak). Grants = HS256 JWT, path-scoped + TTL. ❌ No detection / Web Bot Auth / routing — **machine-only, no human lane** ("routing not gatekeeping" unbuilt). ❌ Origin middleware only (no edge PEP). ❌ Grant is a ~1-hr **access window, not a per-request quota** → no metering decrement. ⚠️ JTI minted but **no replay store** (grant reusable/shareable until TTL). |
| **L4 Pay** | `payments/{provider,mollie,mock,setup}`, `pay_initiate.go`, `webhook.go` | ✅ **Mollie fiat lane built** (real SDK, `metadata.resource_path`, webhook, test-mode complete). Clean provider interface + mock swap. ❌ No x402/USDC agent lane. ❌ No ACP. ⚠️ EUR-only. ⚠️ Per-payment, **not aggregated**. |
| **L5 Aggregate** | `guardrails/budget.go`, `approval/approval.go` | ⚠️ Partial & **on the buyer side**: daily spend cap + Y/N approval live in the **agent wallet**, not seller-side batching. ❌ No settlement batching → **micro-economics unsolved** (each access = 1 Mollie tx; fine at €0.50, breaks at €0.002). |
| **L6 Reconcile** | `server.go` (`payment_id → grant` map), Mollie metadata | ❌ Minimal: payment→grant recognition only. No event log, no ledger, no payout matching, no refund/chargeback/failed-debit handling. **Core moat unbuilt.** |
| **L7 Payout / compliance** | — | ❌ Missing. No payout, PCI scoping, GDPR data handling, AI Act transparency. |

**🔴 Security finding (fix before real money):** `/webhooks/payment` **trusts the payload** — it forces `status=paid` on form posts and only re-fetches the payment to fill *missing metadata*, never to **confirm it was actually paid**. A forged POST with a known/guessable payment id mints a valid access grant → free content. **Fix:** in `OnPaid`, call `provider.GetPayment(id)` and require `status == paid` before `IssueGrant`; add Mollie webhook source/signature verification.

**Gaps to close next (priority order):**
1. **Webhook payment verification** (above) — correctness + security.
2. **Reconciliation event log (L6)** — emit one append-only access/payment event from gateway + webhook; this is the moat and it's currently a `map`.
3. **Asymmetric grant signing** — the shared HMAC secret can't be verified by an independent edge PEP without distributing it; switch to **Ed25519** to make enforcement edge-or-origin-flexible (the L3 differentiation).
4. **Aggregation / batching (L5)** — move spend-cap logic seller-side and batch micro-charges into one settlement so sub-cent pricing is viable.
5. **Detection + routing (L3)** — add Web Bot Auth verify + the crawler/agent/human trichotomy if/when you serve a human lane.
6. **Per-request quota** — make the grant carry a decrementing byte/request quota instead of a flat window.

*Good bones worth keeping: the `PaymentProvider` interface + mock/mollie swap, idempotency-key plumbing, no-leak middleware, and the wallet's budget + approval guardrails.*

---

## USE-CASE TARGETING (where to point the workflow)

| Segment | Pain | Entry move |
|---|---|---|
| **Open newspapers / reference** | bots scrape free; ads earn $0 from bots; blocking loses fastest-growing traffic | per-page micro-fee (e.g. $0.001/page × millions of agent views) |
| **Paywalled news / research** | agents have no subscription; direct deals don't scale to long-tail | RSL encrypted-access + hard gating + aggregation |
| **Premium data / APIs** | enterprise contracts don't scale down | pay-per-record access |
| **Academic / non-profit** | risk of "elitist web" backlash | free/discounted carve-out tier (positioning win) |

---

## COMPETITION — TollBit teardown + the EU data-residency wedge

**Reality check:** TollBit (Novoscribe, Inc., NY) is not "a bot paywall." It has built most of *this entire workflow* at scale — 6,000+ publishers with the demand side already integrated. Do **not** pitch "agentic content paywall" as category creation; the category exists and has a funded leader. Compete on the one seam its architecture can't easily close.

**What TollBit already owns (do NOT claim these as your differentiation):**

| Capability | TollBit status |
|---|---|
| The gate | "Agent Site" subdomain (`tollbit.{domain}`) + edge workers + content cache (not CDN-locked, not origin-bound) |
| Auth | Signed-JWT tokens: crawl / transaction / access; UA validated from token not header; bill only on 200 OK |
| Reconciliation ledger | Immutable, per-tx, price in micros, both-sided — *the "core moat" I'd assigned you, already built* |
| Licensing | Standard (summarization vs full display) + custom/private + $0 ledger-only deals |
| Clean context | DOM strip → markdown, JSON-LD, wire/Getty rights exclusion, embargo/time-decay pricing |
| Demand side | AI cos integrate via proxy / async metering / hybrid discovery; click-through referral rebates |
| Network | 6,000+ publishers; Akamai / Cloudflare / DataDome / HUMAN detection partnerships |

### The one seam: EU data residency

TollBit's design forces structural US exposure that an EU-native product avoids entirely:

1. **It requires publishers to export *all* HTTP logs — including reader IPs (since Q4 2025) — to a US ingestion endpoint.** IPs are personal data (CJEU *Breyer*). That is a transatlantic transfer of reader PII into US surveillance jurisdiction (FISA 702).
2. **US-shaped money/legal:** Stripe Connect payouts, W-9 tax onboarding, USD/CPM pricing.
3. **The legal ground is *valid but fragile*:** the EU–US Data Privacy Framework currently stands (Latombe dismissed at the General Court, Sept 2025) but is under CJEU appeal (Case C-703/25 P, no hearing date as of mid-2026) with a "Schrems III" challenge looming — and two prior frameworks (Safe Harbor, Privacy Shield) were already struck down. A publisher betting compliance on the DPF is betting on something invalidated twice before.

> **Honest framing (survives a general counsel's scrutiny):** *not* "TollBit is illegal" — it isn't, under the DPF — but **"TollBit makes you dependent on a fragile US-transfer framework and exports your readers' IPs to US jurisdiction, when in-region processing needs no transfer mechanism at all."**

**Why this is a real moat, not a checkbox they'd copy:** TollBit's *core value* — cross-network demand detection, analytics, click-through rebates — **structurally depends on centralizing every publisher's logs**. "Keep logs in-region" conflicts with their data model; they can't bolt it on without gutting the analytics that differentiate them. You're attacking a seam, not a feature.

**What you must actually commit to (so the claim is true, not marketing):**
- Process at the **publisher's own edge / in-region**; never exfiltrate raw logs or IPs cross-border.
- **EU data residency** for all storage; EU sub-processors only; pseudonymize/hash IPs at source.
- **EU money + legal:** SEPA/iDEAL/EUR settlement, EU VAT handling, EU-entity DPA, EU AI Act + Data Act alignment.
- A clean, signable **DPA + no-US-transfer attestation** — your actual sales collateral for the DPO.

**GTM consequence:** you sell to the **DPO / general counsel**, not only the revenue team. "No reader data leaves the EU" is a procurement-unblocker no US incumbent can match without re-architecting.

**Target the most compliance-bound verticals** where US players are weakest: **EU scientific/academic, legal, and financial-data publishers** — high-value content, sophisticated buyers, and DPOs who will not sign off on US log export.

### Battlecard — their move → your counter

| TollBit move | Your counter |
|---|---|
| "We have 6,000 publishers" | "None is a [vertical] EU publisher who can sign off on exporting reader IPs to the US." |
| "We added an EU entity / SEPA" | "Payouts moved; your logs and IPs still flow to US ingestion — and that flow is load-bearing for their analytics, so it won't change." |
| "The DPF makes US transfer legal" | "Valid today, struck down twice before, under CJEU appeal now. In-region needs no transfer mechanism at all." |
| "We're the standard" | "The standard is US-shaped. EU publishers need an EU-resident default." |

**The honest risk:** this is a *speed-and-focus* wedge, not a permanent moat. If the category proves out and TollBit re-architects for EU residency, your lead is time, depth in one vertical, and DPO relationships — not technology. **Win the EU vertical before they bother.**

---

## ONE-LINE POSITIONING

> The **GDPR-native, data-resident content-monetization layer for EU research / legal / financial publishers** — gate, meter, and settle agent access entirely in-region, on EU rails (SEPA/iDEAL/EUR), so your logs and your readers' IPs **never leave the EU**.
>
> **Moat = EU data residency at an architectural seam TollBit can't copy without gutting its analytics, plus depth in one compliance-bound vertical. A speed wedge, not permanent — win the vertical first.**

---

## QUICK DECISION CHECKLIST

- [ ] Layer 1 — Ship an MCP server (adopt)
- [ ] Layer 2 — Speak RSL + add enforcement teeth + research carve-out
- [ ] Layer 3 — Build edge-or-origin gate (402→token→quota); Web Bot Auth + behavioral defense; encode the crawler/agent/human trichotomy; **route, don't gatekeep — no CAPTCHA in the bot/token path**
- [ ] Layer 4 — Dual rail: x402/USDC (agents) + Mollie (humans/EU); ACP Delegate-Payment, not SPT
- [ ] Layer 5 — Build aggregation + spend caps (= SCA exemption fit)
- [ ] Layer 6 — Build reconciliation ledger fed by the Layer-3 event stream (core moat)
- [ ] Layer 7 — Offload to Mollie (PSD2/PCI), monitor AI Act Art. 50

*Sources: agentic-commerce protocol coverage (eco.com, Google Developers, Presta, acpready); ACP specs (OpenAI Developers, agenticcommerce.expert); x402 docs (Coinbase/Cloudflare/Alchemy/xpay); Cloudflare Pay-Per-Crawl / AI Crawl Control (Cloudflare blog, ppc.land, tostring.ai dev guide); Web Bot Auth (IETF draft-meunier, Google Developers, Akamai, Cloudflare); Mollie developer docs; EU AI Act guidance (European Commission, Latham, Covington); RSL coverage (RSL Collective, Plagiarism Today, WAN-IFRA, Creative Commons); TollBit platform docs (tollbit.com/docs); competitor landscape (Nieman Lab, Media Copilot, Presenc AI, Akamai); EU–US Data Privacy Framework status (IAPP, Hunton, DLA Piper — Latombe/General Court Sept 2025, CJEU appeal C-703/25 P). Verify all regulatory dates — AI Act omnibus amendments not all formally adopted, and the DPF is under CJEU appeal, as of mid-2026.*
