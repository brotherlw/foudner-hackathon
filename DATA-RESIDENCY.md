# Data Residency & Processing Statement

> **Purpose:** DPO / general-counsel–facing summary of how the Agentic Content Paywall
> handles data, and the machine-readable contract behind the `/.well-known/data-residency`
> endpoint. Fill the `[BRACKETED]` placeholders. **This is a template — have EU privacy
> counsel review before sharing it as a binding statement.**

## Summary

[COMPANY] operates an in-region content-access gateway for EU publishers. By design, the
gateway collects the minimum data needed to authorize and meter a single content access,
stores it within the EU, and performs no transfer of personal data to non-EU jurisdictions.
Unlike log-export monetization models, the gateway does **not** require the publisher to ship
their HTTP access logs (including visitor IP addresses) to a foreign processor.

## What we collect (per gated access event)

| Field | Purpose | Personal data? |
|---|---|---|
| Timestamp | Billing, reconciliation | No |
| Resource path | Which content was accessed | No |
| Amount / currency | Billing | No |
| Agent identity (token claim) | Authorize the paying agent | No (machine identity) |
| Payment ID | Reconcile payment ↔ access | No |
| Decision (granted/denied) | Audit | No |

## What we do NOT collect

- **Raw visitor IP addresses.** Not stored. Where an IP is unavoidably observed in transit,
  it is hashed (SHA-256 with a per-deployment salt) at source and the raw value discarded.
- **Full HTTP access logs** of human traffic. The gateway sees only requests to gated
  resources, not the publisher's general traffic.
- **Reader identity, cookies, behavioral/biometric data**, or any special-category data.

## Where data is stored

- **Region:** [EU REGION, e.g. eu-central / Frankfurt].
- **Storage:** in-region only ([STORAGE, e.g. local append-only ledger / EU object storage]).
- **No replication** to non-EU regions.

## Sub-processors

| Sub-processor | Role | Location | Transfer outside EU? |
|---|---|---|---|
| Mollie | Payment processing | EU (Netherlands) | No |
| [HOSTING] | Compute / hosting | [EU REGION] | No |

No US-based sub-processors. No reliance on the EU–US Data Privacy Framework or Standard
Contractual Clauses, because no personal data leaves the EU.

## Cross-border transfer

**None.** Because processing and storage are in-region and no personal data is transferred to
a third country, no Chapter V (Arts 44–49 GDPR) transfer mechanism is required.

## Legal bases (publisher as controller; [COMPANY] as processor)

- Processing is performed on the publisher's instructions under a Data Processing Agreement.
- Data minimization (Art 5(1)(c)) and storage limitation (Art 5(1)(e)) are enforced by design.
- [Add controller's lawful basis for the underlying processing — to be set by the publisher.]

## Security

- Access grants are cryptographically signed and short-lived.
- Content is not served without a verified grant.
- Payment is confirmed with the PSP before any access is granted.
- Secrets and full tokens are never logged.

## How this differs from log-export monetization

A common competing model requires the publisher to forward all HTTP access logs — including
visitor IP addresses — to a non-EU ingestion endpoint, creating a transatlantic transfer of
reader personal data into a foreign surveillance jurisdiction and a dependency on the
fragile adequacy framework for that jurisdiction. This gateway does neither.

## Verification

The live machine-readable statement is served at `GET /.well-known/data-residency` and is
generated from deployment configuration, so it always reflects the running system. An
immutable access ledger supports audit of every authorization decision.

## Contact

Data Protection contact: [DPO NAME / EMAIL]. DPA available on request.

---
*Template provided for drafting purposes. Not legal advice. Validate with qualified EU
privacy counsel before relying on or distributing this statement.*
