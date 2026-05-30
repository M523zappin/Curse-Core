#requires -Version 5.1
<#
.SYNOPSIS
  CURSE — Zero-Touch Installer (Windows)
.DESCRIPTION
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
function Write-Header($msg) {
    Write-Host "`n  $msg" -ForegroundColor White -BackgroundColor DarkCyan
}

Clear-Host
Write-Host @"

  ╔══════════════════════════════════════════════╗
  ║              C U R S E                       ║
  ║  Zero-Touch Installer (Windows)              ║
  ╚══════════════════════════════════════════════╝

"@ -ForegroundColor Cyan

# ── Dependency: git ──────────────────────────────────────────
Write-Header "Dependencies"
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

# ── Dependency: Go (if no pre-built binary) ─────────────────
$BinaryExists = $false
# Check for Windows binary
$binaryPath = Join-Path $CurseHome 'releases\curse-dashboard.exe'
if ((Test-Path $binaryPath)) { $BinaryExists = $true }

if (-not $BinaryExists -and -not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Step "Installing Go 1.23.4..."
    $goUrl = 'https://go.dev/dl/go1.23.4.windows-amd64.zip'
    $goZip = Join-Path $TempDir 'go.zip'
    New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
    Invoke-WebRequest -Uri $goUrl -OutFile $goZip -UseBasicParsing
    Expand-Archive -Path $goZip -DestinationPath 'C:\Go' -Force
    $env:Path = "C:\Go\bin;$env:Path"
    [Environment]::SetEnvironmentVariable('Path', "C:\Go\bin;$([Environment]::GetEnvironmentVariable('Path','Machine'))", 'Machine')
    [Environment]::SetEnvironmentVariable('GOROOT', 'C:\Go', 'Machine')
    Write-OK 'Go installed'
} else {
    Write-OK 'Go'
}

# ── Clone repository ────────────────────────────────────────
Write-Header "Repository"
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
Write-Header "Binary"
New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
$targetExe = Join-Path $BinDir 'curse.exe'

if (Test-Path (Join-Path $CurseHome 'releases\curse-dashboard.exe')) {
    Copy-Item (Join-Path $CurseHome 'releases\curse-dashboard.exe') $targetExe -Force
    Write-OK 'Pre-built binary deployed'
} elseif (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Step "Building from source..."
    Push-Location $CurseHome
    $env:GOROOT = 'C:\Go'
    $env:Path = "C:\Go\bin;$env:Path"
    go build -o $targetExe ./cmd/dashboard/ 2>&1 | Out-Null
    Pop-Location
    Write-OK "Binary built: $targetExe"
}

# ── Register PATH ───────────────────────────────────────────
Write-Header "Environment"
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notlike "*$BinDir*") {
    [Environment]::SetEnvironmentVariable('Path', "$BinDir;$userPath", 'User')
    $env:Path = "$BinDir;$env:Path"
    Write-OK "Added $BinDir to PATH"
} else {
    Write-OK "PATH already configured"
}

# ── Bootstrap .env ──────────────────────────────────────────
$envPath = Join-Path $CurseHome '.env'
if (-not (Test-Path $envPath)) {
    Copy-Item (Join-Path $CurseHome '.env.example') $envPath
    Write-OK "Created $envPath (edit with your API keys)"
} else {
    Write-OK '.env already exists'
}

# ── GitHub Auth ─────────────────────────────────────────────
Write-Header "GitHub"
if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    Write-Step "Installing GitHub CLI..."
    $ghUrl = 'https://github.com/cli/cli/releases/download/v2.63.2/gh_2.63.2_windows_amd64.msi'
    $ghInstaller = Join-Path $TempDir 'gh.msi'
    Invoke-WebRequest -Uri $ghUrl -OutFile $ghInstaller -UseBasicParsing
    Start-Process msiexec.exe -Wait -ArgumentList "/i $ghInstaller /quiet /norestart"
    $env:Path = "${env:ProgramFiles}\GitHub CLI;$env:Path"
    Write-OK 'GitHub CLI installed'
}

$ghStatus = & gh auth status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Step "Opening GitHub browser auth..."
    Start-Process 'https://github.com/login/device'
    gh auth login --web
}

# ── Cleanup ─────────────────────────────────────────────────
if (Test-Path $TempDir) { Remove-Item $TempDir -Recurse -Force }
"@

Write-Host @"

  ╔══════════════════════════════════════════════╗
  ║              C U R S E                       ║
  ║  Installation Complete                       ║
  ║                                              ║
  ║  Binary: $BinDir\curse.exe"   ║
  ║  Source: $CurseHome           ║
  ║                                              ║
  ║  Run:   curse                                ║
  ╚══════════════════════════════════════════════╝
"@ -ForegroundColor Green
