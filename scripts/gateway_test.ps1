$ErrorActionPreference = "Stop"

$Repo = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $Repo

Write-Host "==> Building gateway"
go build -o bin\gateway.exe .\cmd\gateway

Write-Host "==> Stopping any gateway already listening on :3001"
$listeners = Get-NetTCPConnection -LocalPort 3001 -State Listen -ErrorAction SilentlyContinue |
    Select-Object -ExpandProperty OwningProcess -Unique
foreach ($listener in $listeners) {
    if ($listener -and $listener -ne $PID) {
        Stop-Process -Id $listener -Force -ErrorAction SilentlyContinue
    }
}

Remove-Item .\gateway.err.log, .\gateway.out.log -ErrorAction SilentlyContinue

Write-Host "==> Starting gateway"
$gateway = Join-Path $Repo "bin\gateway.exe"
$stdout = Join-Path $Repo "gateway.out.log"
$stderr = Join-Path $Repo "gateway.err.log"
$process = Start-Process -FilePath $gateway `
    -WorkingDirectory $Repo `
    -RedirectStandardOutput $stdout `
    -RedirectStandardError $stderr `
    -PassThru

Start-Sleep -Seconds 1

try {
    Invoke-WebRequest "http://localhost:3001/.well-known/data-residency" -UseBasicParsing | Out-Null
} catch {
    Write-Host "==> Gateway did not answer. Last stderr lines:"
    if (Test-Path $stderr) {
        Get-Content $stderr -Tail 20
    }
    throw
}

Write-Host "==> Gateway running on http://localhost:3001 with pid $($process.Id)"
Start-Process "http://localhost:3001/demo"

Write-Host "==> Watching gateway stderr log"
Get-Content $stderr -Wait
