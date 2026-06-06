$ErrorActionPreference = "Stop"

$Repo = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Repo

$BaseUrl = if ($env:GATEWAY_URL) { $env:GATEWAY_URL } else { "http://localhost:3001" }
$ResourcePath = if ($env:RESOURCE_PATH) { $env:RESOURCE_PATH } else { "/api/premium-report" }
$Amount = if ($env:AMOUNT) { $env:AMOUNT } else { "0.50" }
$Currency = if ($env:CURRENCY) { $env:CURRENCY } else { "EUR" }

function Write-Step {
    param([string] $Text)
    Write-Host ""
    Write-Host "==> $Text"
}

Write-Step "Checking gateway"
$gatewayCheck = curl.exe -fsS "$BaseUrl/.well-known/data-residency"
if ($LASTEXITCODE -ne 0) {
    throw "gateway is not reachable at $BaseUrl; run scripts\gateway_test.ps1 first"
}
$gatewayCheck
Write-Host ""

Write-Step "Request protected content without grant; expect HTTP 402"
$challengePath = Join-Path $env:TEMP "agentic-paywall-challenge.json"
Remove-Item $challengePath -ErrorAction SilentlyContinue
$status = curl.exe -sS -o $challengePath -w "%{http_code}" "$BaseUrl$ResourcePath"
if ($status -eq "000") {
    throw "gateway is not reachable at $BaseUrl; run scripts\gateway_test.ps1 first"
}
if (Test-Path $challengePath) {
    Get-Content $challengePath
}
Write-Host "HTTP $status"
if ($status -ne "402") {
    throw "expected HTTP 402, got $status"
}

Write-Step "Initiate payment"
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
if (-not $grant.access_grant) {
    throw "access_grant missing from /grants/verify response"
}
Write-Host "ready=$($grant.ready)"
Write-Host "grant=$($grant.access_grant.Substring(0, 24))..."

Write-Step "Retry protected content with PAYMENT-GRANT; expect HTTP 200"
curl.exe -fsS "$BaseUrl$ResourcePath" -H "PAYMENT-GRANT: $($grant.access_grant)"
Write-Host ""

Write-Step "Reuse same grant; expect HTTP 402 after quota"
$quotaPath = Join-Path $env:TEMP "agentic-paywall-quota.json"
$quotaStatus = curl.exe -sS -o $quotaPath -w "%{http_code}" "$BaseUrl$ResourcePath" -H "PAYMENT-GRANT: $($grant.access_grant)"
Get-Content $quotaPath
Write-Host "HTTP $quotaStatus"
if ($quotaStatus -ne "402") {
    throw "expected HTTP 402 after quota, got $quotaStatus"
}

Write-Step "Agent test complete"
