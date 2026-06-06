# Demo Runbook

This runbook shows the Sprint 3 loop:

agent request -> HTTP 402 -> EUR payment -> test completion -> verified grant -> content -> in-region ledger.

## Prerequisites

- Go installed.
- Gateway config copied from `config.json.example` to `config.json`.
- For Mollie test mode: `MOLLIE_API_KEY=test_...`.

Default local demo uses `provider=mock`. To use Mollie test mode, set:

```powershell
set PAYMENT_PROVIDER=mollie
set MOLLIE_API_KEY=test_your_key_here
```

The demo completion endpoint is controlled by:

```json
"demo": {
  "enable_test_completion": true
}
```

Disable it outside local/test demos.

## Start Gateway

Windows CMD:

```cmd
cd /d C:\Users\javie\Documents\Codex\2026-06-06\files-mentioned-by-the-user-pasted\work\sandbox\repo\foudner-hackathon
go build -o bin\gateway.exe .\cmd\gateway
bin\gateway.exe
```

Linux/macOS:

```sh
go build -o bin/gateway ./cmd/gateway
./bin/gateway
```

Leave the gateway running.

## Run Demo Script

In a second terminal on Linux/macOS or Git Bash:

```sh
./scripts/demo.sh
```

Windows PowerShell:

```powershell
.\scripts\demo.ps1
```

For Mollie:

```sh
PAYMENT_PROVIDER=mollie MOLLIE_API_KEY=test_your_key_here ./bin/gateway
./scripts/demo.sh
```

Windows PowerShell with Mollie:

```powershell
$env:PAYMENT_PROVIDER = "mollie"
$env:MOLLIE_API_KEY = "test_your_key_here"
bin\gateway.exe
```

Then in a second PowerShell:

```powershell
.\scripts\demo.ps1
```

## Windows Manual Commands

PowerShell:

```powershell
curl.exe http://localhost:3001/.well-known/data-residency
curl.exe -i http://localhost:3001/api/premium-report
$initiateBody = @{ resource_path = "/api/premium-report"; amount = "0.50"; currency = "EUR" } | ConvertTo-Json -Compress
$initiateBody | Set-Content initiate-payment.json -NoNewline
$payment = curl.exe -X POST http://localhost:3001/pay/initiate -H "Content-Type: application/json" --data-binary "@initiate-payment.json" | ConvertFrom-Json
$completeBody = @{ payment_id = $payment.payment_id } | ConvertTo-Json -Compress
$completeBody | Set-Content complete-test.json -NoNewline
curl.exe -X POST http://localhost:3001/pay/complete-test -H "Content-Type: application/json" --data-binary "@complete-test.json"
$grant = (curl.exe "http://localhost:3001/grants/verify?payment_id=$($payment.payment_id)" | ConvertFrom-Json).access_grant
curl.exe http://localhost:3001/api/premium-report -H "PAYMENT-GRANT: $grant"
Get-Content .\ledger\events.jsonl
```

## Expected Proof Points

- First content request returns HTTP 402.
- Payment is in EUR.
- `/pay/complete-test` completes the mock/Mollie test payment.
- `/grants/verify` returns an access grant.
- Retried content request returns the premium report.
- `ledger/events.jsonl` contains payment and access events.
- `/.well-known/data-residency` says `raw_ip_retained:false` and `cross_border_transfer:false`.
