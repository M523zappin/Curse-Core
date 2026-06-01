# CURSE - Windows Installer (PowerShell)
# Run this in PowerShell:

$ErrorActionPreference = "Stop"

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

# Check Windows version
if (-not ($env:OS -eq "Windows_NT")) {
    Write-Host "Error: This requires Windows" -ForegroundColor $RED
    exit 1
}

$BIN_DIR = "$env:USERPROFILE\AppData\Local\Bin"
$INSTALL_PATH = "$BIN_DIR\curse.exe"

# Create directory if needed
if (-not (Test-Path $BIN_DIR)) {
    New-Item -ItemType Directory -Path $BIN_DIR -Force | Out-Null
}

Write-Host "  Platform: Windows" -ForegroundColor Cyan
Write-Host "  Install to: $INSTALL_PATH" -ForegroundColor Cyan
Write-Host ""

$REPO = "M523zappin/Curse-Core"
$API_URL = "https://api.github.com/repos/$REPO/releases/latest"

try {
    # Get latest release
    Write-Host "  Fetching latest version..." -ForegroundColor Cyan
    $RELEASE = Invoke-RestMethod $API_URL -UseBasicParsing -TimeoutSec 10
    $VERSION = $RELEASE.tag_name
    $FILENAME = "curse-windows-amd64.exe"
    $DOWNLOAD_URL = "https://github.com/$REPO/releases/download/$VERSION/$FILENAME"
    
    Write-Host "  Downloading v$VERSION..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $INSTALL_PATH -UseBasicParsing -TimeoutSec 60
    
    # Remove Zone identifier if present
    Unblock-File -Path $INSTALL_PATH -ErrorAction SilentlyContinue
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  INSTALLED SUCCESSFULLY!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    
} catch {
    Write-Host "  Download failed: $_" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Yellow
    Write-Host "  ALTERNATIVE: Build from source" -ForegroundColor Yellow
    Write-Host "========================================" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  1. Install Go from: https://go.dev/dl/" -ForegroundColor Cyan
    Write-Host "  2. Restart PowerShell" -ForegroundColor Cyan  
    Write-Host "  3. Run: go install github.com/$REPO/cmd/dashboard@latest" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  Or use pre-built binaries from:" -ForegroundColor Yellow
    Write-Host "  https://github.com/$REPO/releases" -ForegroundColor Cyan
    Write-Host ""
    exit 1
}

# Add to PATH
$USER_PATH = [Environment]::GetEnvironmentVariable("Path", "User")
if ($USER_PATH -notlike "*$BIN_DIR*") {
    [Environment]::SetEnvironmentVariable("Path", "$USER_PATH;$BIN_DIR", "User")
    $env:Path = "$USER_PATH;$BIN_DIR"
}

Write-Host "  Quick Start:" -ForegroundColor Cyan
Write-Host "    curse" -ForegroundColor White
Write-Host ""
Write-Host "  Examples:" -ForegroundColor Cyan
Write-Host '    >>> create a REST API in Go' -ForegroundColor White
Write-Host '    >>> add unit tests' -ForegroundColor White
Write-Host '    >>> write a Dockerfile' -ForegroundColor White
Write-Host ""
