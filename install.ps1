#requires -Version 5.1
<#
.SYNOPSIS
  CURSE — Autonomous Installer (Windows)
.DESCRIPTION
  No API keys needed. No forced cloud auth. Just run.
  Run:  iex "& { $(irm https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.ps1) }"
#>

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repo = 'M523zappin/Curse-Core'
$Branch = 'master'
$CurseHome = Join-Path $env:USERPROFILE 'curse'
$BinDir = Join-Path $env:USERPROFILE '.local\bin'
$TempDir = Join-Path $env:TEMP 'curse-install'

function Write-Step($msg) { Write-Host "  → $msg" -ForegroundColor Cyan }
function Write-OK($msg)   { Write-Host "  ✔ $msg" -ForegroundColor Green }

Clear-Host
Write-Host @"

  ╔══════════════════════════════════════════════╗
  ║              C U R S E                       ║
  ║  Autonomous Installer — Zero API Keys        ║
  ╚══════════════════════════════════════════════╝

"@ -ForegroundColor Cyan

# ── Dependency: git ──────────────────────────────────────────
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Step "Installing Git for Windows..."
    $gitUrl = 'https://github.com/git-for-windows/git/releases/latest/download/Git-2.48.1-64-bit.exe'
    $gitInstaller = Join-Path $TempDir 'git-installer.exe'
    New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
    Invoke-WebRequest -Uri $gitUrl -OutFile $gitInstaller -UseBasicParsing
    Start-Process -Wait -FilePath $gitInstaller -ArgumentList '/VERYSILENT /NORESTART /NOCANCEL /SP- /SUPPRESSMSGBOXES /DIR="C:\Program Files\Git"'
    $env:Path = "C:\Program Files\Git\cmd;$env:Path"
    Write-OK 'Git installed'
} else {
    Write-OK 'Git'
}

# ── Clone repository ────────────────────────────────────────
if (Test-Path (Join-Path $CurseHome '.git')) {
    Write-Step "Updating existing installation at $CurseHome"
    Push-Location $CurseHome
    git pull --ff-only origin $Branch 2>&1 | Out-Null
    Pop-Location
} else {
    Write-Step "Cloning $Repo → $CurseHome"
    if (Test-Path $CurseHome) { Remove-Item $CurseHome -Recurse -Force }
    git clone --depth 1 --branch $Branch "https://github.com/$Repo.git" $CurseHome
}
Write-OK "Repository at $CurseHome"

# ── Build / copy binary ─────────────────────────────────────
New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
$targetExe = Join-Path $BinDir 'curse.exe'

# Try pre-built binary first (search multiple locations)
$prebuiltPaths = @(
    Join-Path $CurseHome 'curse.exe'
    Join-Path $CurseHome 'releases\curse-dashboard.exe'
    Join-Path $CurseHome 'curse-dashboard.exe'
)
$deployed = $false
foreach ($src in $prebuiltPaths) {
    if (Test-Path $src) {
        Copy-Item $src $targetExe -Force
        Write-OK "Pre-built binary deployed: $src"
        $deployed = $true
        break
    }
}

if (-not $deployed) {
    # Build from source
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Step "Installing Go 1.26.0..."
        $goUrl = 'https://go.dev/dl/go1.26.0.windows-amd64.zip'
        $goZip = Join-Path $TempDir 'go.zip'
        New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
        Invoke-WebRequest -Uri $goUrl -OutFile $goZip -UseBasicParsing
        Expand-Archive -Path $goZip -DestinationPath 'C:\Go' -Force
        $env:Path = "C:\Go\bin;$env:Path"
        [Environment]::SetEnvironmentVariable('Path', "C:\Go\bin;$([Environment]::GetEnvironmentVariable('Path','Machine'))", 'Machine')
        [Environment]::SetEnvironmentVariable('GOROOT', 'C:\Go', 'Machine')
        Write-OK 'Go installed'
    }
    Write-Step "Building from source..."
    Push-Location $CurseHome
    $env:GOROOT = 'C:\Go'
    $env:Path = "C:\Go\bin;$env:Path"
    go build -o $targetExe ./cmd/dashboard/ 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Pop-Location
        Write-Host "Build failed. Check Go installation." -ForegroundColor Red
        exit 1
    }
    Pop-Location
    Write-OK "Binary built: $targetExe"
}

# Verify the binary works
if (-not (Test-Path $targetExe)) {
    Write-Host "Binary not found at $targetExe. Install failed." -ForegroundColor Red
    exit 1
}

# ── Register PATH ───────────────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable('Path', "$BinDir;$userPath", 'User')
    $env:Path = "$BinDir;$env:Path"
    Write-OK "Added $BinDir to PATH"
} else {
    Write-OK "PATH already configured"
}

# ── Cleanup ─────────────────────────────────────────────────
if (Test-Path $TempDir) { Remove-Item $TempDir -Recurse -Force }

Write-Host @"

  ╔══════════════════════════════════════════════╗
  ║              C U R S E                       ║
  ║  Installation Complete                       ║
  ║                                              ║
  ║  No API keys needed. Just run.               ║
  ║                                              ║
  ║  Binary: $BinDir\curse.exe
  ║  Source: $CurseHome
  ║                                              ║
  ║  Run:   curse                                ║
  ╚══════════════════════════════════════════════╝
"@ -ForegroundColor Green
