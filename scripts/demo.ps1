$ErrorActionPreference = "Stop"

$BaseUrl = if ($env:GATEWAY_URL) { $env:GATEWAY_URL } else { "http://localhost:3001" }
$ResourcePath = if ($env:RESOURCE_PATH) { $env:RESOURCE_PATH } else { "/api/premium-report" }
$Amount = if ($env:AMOUNT) { $env:AMOUNT } else { "0.50" }
$Currency = if ($env:CURRENCY) { $env:CURRENCY } else { "EUR" }
$LedgerPath = if ($env:LEDGER_PATH) { $env:LEDGER_PATH } else { "ledger\events.jsonl" }

function Write-Step {
    param([string] $Text)
    Write-Host ""
    Write-Host "==> $Text"
}

Write-Step "Data residency statement"
curl.exe -fsS "$BaseUrl/.well-known/data-residency"
Write-Host ""

Write-Step "Request protected content without a grant; expect HTTP 402"
$challengePath = Join-Path $env:TEMP "agentic-paywall-challenge.json"
$status = curl.exe -sS -o $challengePath -w "%{http_code}" "$BaseUrl$ResourcePath"
Get-Content $challengePath
Write-Host "HTTP $status"
if ($status -ne "402") {
    throw "expected HTTP 402, got $status"
}

Write-Step "Initiate EUR payment"
$initiateBody = @{
    resource_path = $ResourcePath
    amount = $Amount
    currency = $Currency
} | ConvertTo-Json -Compress
$initiatePath = Join-Path $env:TEMP "agentic-paywall-initiate.json"
Set-Content -Path $initiatePath -Value $initiateBody -NoNewline
$payment = curl.exe -fsS -X POST "$BaseUrl/pay/initiate" -H "Content-Type: application/json" --data-binary "@$initiatePath" | ConvertFrom-Json
$payment | ConvertTo-Json -Compress
if (-not $payment.payment_id) {
    throw "payment_id missing from /pay/initiate response"
}
if ($payment.checkout_url) {
    Write-Host "Checkout URL: $($payment.checkout_url)"
}

Write-Step "Complete test payment"
$completeBody = @{
    payment_id = $payment.payment_id
} | ConvertTo-Json -Compress
$completePath = Join-Path $env:TEMP "agentic-paywall-complete.json"
Set-Content -Path $completePath -Value $completeBody -NoNewline
curl.exe -fsS -X POST "$BaseUrl/pay/complete-test" -H "Content-Type: application/json" --data-binary "@$completePath"
Write-Host ""

Write-Step "Verify grant"
$grant = curl.exe -fsS "$BaseUrl/grants/verify?payment_id=$($payment.payment_id)" | ConvertFrom-Json
$grant | ConvertTo-Json -Compress
if (-not $grant.access_grant) {
    throw "access_grant missing from /grants/verify response"
}

Write-Step "Retry protected content with PAYMENT-GRANT"
curl.exe -fsS "$BaseUrl$ResourcePath" -H "PAYMENT-GRANT: $($grant.access_grant)"
Write-Host ""

Write-Step "Ledger"
if (-not (Test-Path $LedgerPath)) {
    throw "ledger not found at $LedgerPath"
}
Get-Content $LedgerPath
