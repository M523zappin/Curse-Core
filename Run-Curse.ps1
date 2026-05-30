#!/usr/bin/env pwsh
# Run-Curse.ps1 — Single command to launch the Curse Gateway TUI dashboard
# Also optionally starts a demo mission feed for live testing

param(
    [switch]$Demo,
    [switch]$Help
)

if ($Help) {
    @"
Curse Gateway — TUI Dashboard Launcher
Usage:
  .\curse.cmd          Launch the dashboard (Windows batch)
  .\Run-Curse.ps1      Launch the dashboard
  .\Run-Curse.ps1 -Demo  Launch dashboard with a demo mission feed

The dashboard connects to:
  ~/.curse/logs/event.log     — Live event stream (polled every 500ms)
  ~/.curse/logs/session.json  — Checkpoint state
  %APPDATA%/curse/models.json — Model profiles
"@
    exit 0
}

$curseDir = "$env:USERPROFILE\.curse"
$logDir = "$curseDir\logs"

# Ensure directories exist
if (-not (Test-Path $logDir)) { New-Item -ItemType Directory -Path $logDir -Force | Out-Null }

# Start demo feed if requested
if ($Demo) {
    $logPath = "$logDir\event.log"
    $cpPath = "$logDir\session.json"

    # Pre-populate event.log with sample events
    @"
{"id":"demo-0001","sequence":1,"prev_state":0,"event":0,"new_state":1,"timestamp":"$(Get-Date -Format 'o')","data":{"step":"0","mission":"demo-refactor","event":"MissionStarted"},"checksum":"a1b2c3d4e5f6"}
{"id":"demo-0002","sequence":2,"prev_state":1,"event":1,"new_state":1,"timestamp":"$(Get-Date -Format 'o')","data":{"step":"1","mission":"demo-refactor","event":"StepCompleted"},"checksum":"b2c3d4e5f6a7"}
{"id":"demo-0003","sequence":3,"prev_state":1,"event":1,"new_state":1,"timestamp":"$(Get-Date -Format 'o')","data":{"step":"2","mission":"demo-refactor","event":"StepCompleted"},"checksum":"c3d4e5f6a7b8"}
{"id":"demo-0004","sequence":4,"prev_state":1,"event":1,"new_state":1,"timestamp":"$(Get-Date -Format 'o')","data":{"step":"3","mission":"demo-refactor","event":"StepCompleted"},"checksum":"d4e5f6a7b8c9"}
{"id":"demo-0005","sequence":5,"prev_state":1,"event":2,"new_state":3,"timestamp":"$(Get-Date -Format 'o')","data":{"step":"0","mission":"demo-refactor","event":"CheckpointDue"},"checksum":"e5f6a7b8c9d0"}
{"id":"demo-0006","sequence":6,"prev_state":3,"event":3,"new_state":1,"timestamp":"$(Get-Date -Format 'o')","data":{"step":"0","mission":"demo-refactor","event":"CheckpointWritten"},"checksum":"f6a7b8c9d0e1"}
"@ | Out-File -FilePath $logPath -Encoding utf8

    # Pre-populate session.json
    @"
{
  "state": 1,
  "step": 0,
  "mission_id": "demo-refactor",
  "sequence": 6,
  "last_hash": "f6a7b8c9d0e1",
  "staged_files": ["internal/sandbox/cache.go"],
  "timestamp": "$(Get-Date -Format 'o')"
}
"@ | Out-File -FilePath $cpPath -Encoding utf8

    Write-Host "Demo feed initialized: 6 events, 1 checkpoint" -ForegroundColor Cyan
}

Write-Host "Launching Curse Gateway TUI..." -ForegroundColor Green
Write-Host "  Ctrl+P  Pause/Resume" -ForegroundColor DarkGray
Write-Host "  Ctrl+M  Cycle model" -ForegroundColor DarkGray
Write-Host "  Ctrl+S  Shutdown" -ForegroundColor DarkGray
Write-Host ""

$exe = Join-Path $PSScriptRoot "curse-dashboard.exe"
if (-not (Test-Path $exe)) {
    Write-Host "Building dashboard..." -ForegroundColor Yellow
    $env:GOROOT = "C:\Go"
    $env:Path = "C:\Go\bin;$env:Path"
    & "C:\Go\bin\go.exe" build -o $exe ./cmd/dashboard/
}

& $exe
