#!/usr/bin/env sh
set -eu

BASE_URL="${GATEWAY_URL:-http://localhost:3001}"
RESOURCE_PATH="${RESOURCE_PATH:-/api/premium-report}"
AMOUNT="${AMOUNT:-0.50}"
CURRENCY="${CURRENCY:-EUR}"
LEDGER_PATH="${LEDGER_PATH:-ledger/events.jsonl}"

json_field() {
  python3 -c 'import json,sys; print(json.load(sys.stdin).get(sys.argv[1], ""))' "$1"
}

say() {
  printf '\n==> %s\n' "$1"
}

say "Data residency statement"
curl -fsS "$BASE_URL/.well-known/data-residency"
printf '\n'

say "Request protected content without a grant; expect HTTP 402"
status="$(curl -sS -o /tmp/agentic-paywall-challenge.json -w "%{http_code}" "$BASE_URL$RESOURCE_PATH")"
cat /tmp/agentic-paywall-challenge.json
printf '\nHTTP %s\n' "$status"
if [ "$status" != "402" ]; then
  echo "expected HTTP 402, got $status" >&2
  exit 1
fi

say "Initiate EUR payment"
initiate_body="$(printf '{"resource_path":"%s","amount":"%s","currency":"%s"}' "$RESOURCE_PATH" "$AMOUNT" "$CURRENCY")"
initiate="$(curl -fsS -X POST "$BASE_URL/pay/initiate" \
  -H "Content-Type: application/json" \
  -d "$initiate_body")"
printf '%s\n' "$initiate"
payment_id="$(printf '%s' "$initiate" | json_field payment_id)"
checkout_url="$(printf '%s' "$initiate" | json_field checkout_url)"
if [ -z "$payment_id" ]; then
  echo "payment_id missing from /pay/initiate response" >&2
  exit 1
fi
if [ -n "$checkout_url" ]; then
  printf 'Checkout URL: %s\n' "$checkout_url"
fi

say "Complete test payment"
complete_body="$(printf '{"payment_id":"%s"}' "$payment_id")"
curl -fsS -X POST "$BASE_URL/pay/complete-test" \
  -H "Content-Type: application/json" \
  -d "$complete_body"
printf '\n'

say "Verify grant"
grant_json="$(curl -fsS "$BASE_URL/grants/verify?payment_id=$payment_id")"
printf '%s\n' "$grant_json"
access_grant="$(printf '%s' "$grant_json" | json_field access_grant)"
if [ -z "$access_grant" ]; then
  echo "access_grant missing from /grants/verify response" >&2
  exit 1
fi

say "Retry protected content with PAYMENT-GRANT"
curl -fsS "$BASE_URL$RESOURCE_PATH" -H "PAYMENT-GRANT: $access_grant"
printf '\n'

say "Ledger"
if [ -f "$LEDGER_PATH" ]; then
  cat "$LEDGER_PATH"
else
  echo "ledger not found at $LEDGER_PATH" >&2
  exit 1
fi

