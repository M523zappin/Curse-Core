@echo off
title Curse Gateway — TUI Dashboard
cd /d "%~dp0"
if exist "releases\curse-dashboard.exe" (
    start /b /wait releases\curse-dashboard.exe
) else (
    start /b /wait curse-dashboard.exe
)
