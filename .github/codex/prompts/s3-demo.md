# Sprint 3 — End-to-end EUR demo loop (Mollie test mode)
# Run: codex exec --sandbox workspace-write "$(cat .github/codex/prompts/s3-demo.md)"

Make the end-to-end loop runnable and reproducible with the Mollie test provider in EUR.
Verify the config.provider=mollie path works: agent requests a protected resource -> 402 with the
AgentPaywall challenge -> /pay/initiate creates a Mollie test payment -> CompleteTestPayment
(changePaymentState) -> webhook verifies paid (Sprint 1) -> grant issued -> retry with grant ->
200 + content, with every step appended to the in-region ledger (Sprint 2). Fix any wiring gaps.
Add scripts/demo.sh that runs the full loop against a locally running gateway and prints each step
clearly, and a DEMO.md runbook explaining how to run it. Do NOT add paid dependencies or any non-EU
service. Follow AGENTS.md. Keep build, vet, and test green.
