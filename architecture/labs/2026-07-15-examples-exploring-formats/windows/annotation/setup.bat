@echo off
:: examples/windows/setup.bat (Annotation Syntax)
::
:: Ouvrage Calque - Windows Batch Template
::
:: This file is a fully valid and executable Windows Batch script in local dev,
:: which bypasses the macro definition using standard Batch labels and GOTO.

goto :skip_macros

:: ==========================================
:: 1. MACRO DEFINITION (BATCH LABELS WRAPPING)
:: ==========================================
:: @ocq:Macro(name="ping_db", args=["host"])
:: @ocq:Strip(lines=1)
:define_ping_db
    :: In local dev, we write a valid ping command. AOT replaces it with template variable.
    :: @ocq:Replace(with="    ping -n 1 ${host}")
    ping -n 1 127.0.0.1
    
    :: @ocq:Strip(lines=1)
    exit /b
:: @ocq:EndMacro

:skip_macros

echo --- Starting Windows Environment Initialization ---

:: ==========================================
:: 2. RUNTIME ACTIVE SHADOW MOCK
:: ==========================================
:: @ocq:If(cond=(env.NAME == "production"))
    :: @ocq:Call(macro="ping_db", host="db-prod.siemens.net")
:: @ocq:Else
    echo Local dev: fallback to local database ping simulation.
    ping -n 1 127.0.0.1 >nul
:: @ocq:EndIf

:: ==========================================
:: 3. RUNTIME FILE SYSTEM INSPECTION (DELAYED EXPANSION & GEOMETRY)
:: ==========================================
setlocal enabledelayedexpansion
set LOG_COUNT=0

echo Scanning runtime log directory...
for /f "tokens=*" %%f in ('dir /b *.log 2^>nul') do (
    set /a LOG_COUNT+=1
    
    :: @ocq:Indent(delta=4)
    :: @ocq:Replace(with="    echo [!LOG_COUNT!] Log %%f verified by SRE Team: ${env.TEAM_OWNER}")
    echo [!LOG_COUNT!] Log %%f verified by SRE Team: Siemens-LocalDev-Team
)

echo --- Setup Complete ---
