# CURSE - Windows Installer (PowerShell)
# Just run this ONE command in PowerShell:
#   iex "& { $(irm https://curse.sh/install.ps1) }"

$CYAN = "Cyan"
$GREEN = "Green"
$RED = "Red"
$YELLOW = "Yellow"

Write-Host ""
Write-Host "╔══════════════════════════════════════════╗" -ForegroundColor $CYAN
Write-Host "║         C U R S E  Installer             ║" -ForegroundColor $CYAN
Write-Host "║    Zero API Keys • 100% Offline Ready   ║" -ForegroundColor $CYAN
Write-Host "╚══════════════════════════════════════════╝" -ForegroundColor $CYAN
Write-Host ""

# Detect OS
$OS = [Environment]::OSVersion.Platform
if ($OS -ne "Win32NT") {
    Write-Host "Error: This script requires Windows 10+" -ForegroundColor $RED
    exit 1
}

Write-Host "  * Platform: Windows" -ForegroundColor Cyan

# Install location
$BIN_DIR = "$env:USERPROFILE\AppData\Local\Bin"
if (-not (Test-Path $BIN_DIR)) {
    New-Item -ItemType Directory -Path $BIN_DIR -Force | Out-Null
}
$INSTALL_PATH = "$BIN_DIR\curse.exe"

Write-Host "  * Installing to: $INSTALL_PATH" -ForegroundColor Cyan

# Try to download from GitHub releases
$REPO = "M523zappin/Curse-Core"
$API_URL = "https://api.github.com/repos/$REPO/releases/latest"
try {
    $VERSION = (Invoke-RestMethod $API_URL -UseBasicParsing).tag_name
    $FILENAME = "curse-windows-amd64.exe"
    $DOWNLOAD_URL = "https://github.com/$REPO/releases/download/$VERSION/$FILENAME"
    
    Write-Host "  * Downloading CURSE..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $INSTALL_PATH -UseBasicParsing
    
    Write-Host "  [OK] Installed!" -ForegroundColor Green
} catch {
    Write-Host "  [!] Binary not available, building from source..." -ForegroundColor Yellow
    
    # Check for Go
    $GO = Get-Command go -ErrorAction SilentlyContinue
    if ($GO) {
        Write-Host "  * Building with Go..." -ForegroundColor Cyan
        & go install github.com/$REPO/cmd/dashboard@latest
        $INSTALL_PATH = "$env:GOPATH\bin\dashboard.exe"
    } else {
        Write-Host "  [!] Go not found. Install Go from https://go.dev then run:" -ForegroundColor Yellow
        Write-Host "      go install github.com/$REPO/cmd/dashboard@latest" -ForegroundColor Cyan
    }
}

# Add to PATH
$USER_PATH = [Environment]::GetEnvironmentVariable("Path", "User")
if ($USER_PATH -notlike "*$BIN_DIR*") {
    [Environment]::SetEnvironmentVariable("Path", "$USER_PATH;$BIN_DIR", "User")
    Write-Host "  [!] Added $BIN_DIR to PATH" -ForegroundColor Yellow
    Write-Host "      Restart your terminal or run: refreshenv" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "[OK] CURSE installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "  Quick Start:" -ForegroundColor Cyan
Write-Host "    curse              # Start CURSE"
Write-Host "    curse --help      # Show help"
Write-Host ""
Write-Host "  Examples:" -ForegroundColor Cyan
Write-Host '    >>> create a REST API in Go'
Write-Host '    >>> add authentication middleware'
Write-Host '    >>> write unit tests'
Write-Host ""
