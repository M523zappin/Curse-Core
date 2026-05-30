@echo off
title CURSE — Cognitive Unified Runtime System Entity

:: Find the curse binary in priority order:
::   1. Same directory as this script
::   2. USERPROFILE\.local\bin (install script target)
::   3. PATH

set "CURSE_EXE=%~dp0curse.exe"
if not exist "%CURSE_EXE%" set "CURSE_EXE=%USERPROFILE%\.local\bin\curse.exe"
if not exist "%CURSE_EXE%" set "CURSE_EXE=curse.exe"

start /b /wait "" "%CURSE_EXE%" %*
