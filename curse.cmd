@echo off
title CURSE — Cognitive Unified Runtime System Entity
cd /d "%~dp0"

where /q curse.exe 2>nul
if %ERRORLEVEL% EQU 0 (
    curse.exe %*
    exit /b
)

if exist "releases\curse-dashboard.exe" (
    start /b /wait releases\curse-dashboard.exe
) else (
    start /b /wait curse-dashboard.exe
)
