[CmdletBinding()]
param(
    [switch]$Build
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

function Stop-WithMessage([string]$Message) {
    Write-Error $Message
    exit 1
}

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Stop-WithMessage "Docker Desktop is not installed or docker is not on PATH."
}

& docker info *> $null
if ($LASTEXITCODE -ne 0) {
    Stop-WithMessage "Docker Desktop is not running. Start Docker Desktop and retry."
}

$envPath = Join-Path $repoRoot ".env"
if (-not (Test-Path $envPath)) {
    if (-not (Test-Path (Join-Path $repoRoot ".env.example"))) {
        Stop-WithMessage "Missing .env.example."
    }
    Copy-Item (Join-Path $repoRoot ".env.example") $envPath
    Write-Warning "Created .env from .env.example. Review credentials before production use."
}

# Build the frontend only when the production bundle is absent.
$webIndex = Join-Path $repoRoot "web\dist\index.html"
if (-not (Test-Path $webIndex)) {
    if (-not (Get-Command npm -ErrorAction SilentlyContinue)) {
        Stop-WithMessage "web/dist is missing and npm is not available. Install Node.js, then retry."
    }
    Push-Location (Join-Path $repoRoot "web")
    try {
        if (-not (Test-Path "node_modules")) {
            & npm ci
            if ($LASTEXITCODE -ne 0) { Stop-WithMessage "npm ci failed." }
        }
        & npm run build
        if ($LASTEXITCODE -ne 0) { Stop-WithMessage "npm run build failed." }
        & npm run build:mini
        if ($LASTEXITCODE -ne 0) { Stop-WithMessage "npm run build:mini failed." }
    } finally {
        Pop-Location
    }
}

$composePrefix = @("compose", "--env-file", ".env")
$services = @("postgres", "redis", "migrate", "api", "nginx", "worker")
$upArgs = $composePrefix + @("up", "-d")
if ($Build) { $upArgs += "--build" }
$upArgs += $services

Write-Host "Starting Sola infrastructure and backend..." -ForegroundColor Cyan
& docker @upArgs
if ($LASTEXITCODE -ne 0) {
    Stop-WithMessage "Docker Compose failed while starting the backend."
}

$tokenLine = Get-Content $envPath | Where-Object { $_ -match "^SOLA_BOT_TOKEN=" } | Select-Object -First 1
$token = ""
if ($null -ne $tokenLine) {
    $token = ($tokenLine -replace "^SOLA_BOT_TOKEN=", "").Trim().Trim('"').Trim("'")
}
$hasToken = -not [string]::IsNullOrWhiteSpace($token) -and $token -notmatch "^(replace-with|change-this|your-|placeholder)"

if ($hasToken) {
    Write-Host "Starting Telegram Bot..." -ForegroundColor Cyan
    & docker @($composePrefix + @("up", "-d", "bot"))
    if ($LASTEXITCODE -ne 0) {
        Stop-WithMessage "Backend started, but the Telegram Bot failed to start. Check: docker compose logs bot"
    }
} else {
    Write-Warning "SOLA_BOT_TOKEN is not configured. Backend is running, Telegram Bot was not started."
}

$portLine = Get-Content $envPath | Where-Object { $_ -match "^SOLA_HTTP_PORT=" } | Select-Object -First 1
$port = "80"
if ($null -ne $portLine) {
    $configuredPort = ($portLine -replace "^SOLA_HTTP_PORT=", "").Trim()
    if ($configuredPort) { $port = $configuredPort }
}

Write-Host ""
Write-Host "Sola is running." -ForegroundColor Green
Write-Host "Dashboard: http://localhost:$port/"
Write-Host "Check status: docker compose --env-file .env ps"
if (-not $hasToken) {
    Write-Host "Set SOLA_BOT_TOKEN in .env, then run this script again to start Telegram Bot."
}
