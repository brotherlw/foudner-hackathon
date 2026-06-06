$ErrorActionPreference = "Stop"

$Repo = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Repo

$ledger = Join-Path $Repo "ledger\events.jsonl"
if (!(Test-Path $ledger)) {
    New-Item -ItemType File -Force $ledger | Out-Null
}

Write-Host "==> Watching customer ledger events"
Write-Host "==> $ledger"
Get-Content $ledger -Tail 0 -Wait
