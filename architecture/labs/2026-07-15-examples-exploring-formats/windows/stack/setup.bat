@echo off
:: examples/windows/setup.bat
::
:: Ouvrage Calque - Windows Batch Template
::
:: This file is a fully valid and executable Windows Batch script in local dev,
:: which bypasses the macro definition using standard Batch labels and GOTO.

goto :skip_macros

:: ==========================================
:: 1. MACRO DEFINITION (BATCH LABELS WRAPPING)
:: ==========================================
:: @ocalque:stdlib:macro:define<ping_db>
:: @ocalque:stdlib:macro:arg<ping_db> name="host" type="string" required="true"
:: @ocalque:geometry:strip direction="next"
:define_ping_db
    :: @ocalque:geometry:indent:pushd value="{{ macro.self.ref.indent - self.indent }}"
    
    :: In local dev, we write a valid ping command. AOT replaces it with template variable.
    :: @ocalque:replace_line --- ping -n 1 {{ args.host }}
    ping -n 1 127.0.0.1
    
    :: @ocalque:geometry:indent:popd
    :: @ocalque:geometry:strip direction="next"
    exit /b
:: @ocalque:stdlib:macro:end<ping_db>

:skip_macros

echo --- Starting Windows Environment Initialization ---

:: ==========================================
:: 2. RUNTIME ACTIVE SHADOW MOCK
:: ==========================================
:: @ocalque:stdlib:if<prod_only> --- "{{ env.NAME }}" == "production"
    :: @ocalque:stdlib:macro:call<ping_db> host="db-prod.siemens.net"
:: @ocalque:stdlib:else<prod_only>
    echo Local dev: fallback to local database ping simulation.
    ping -n 1 127.0.0.1 >nul
:: @ocalque:stdlib:fi<prod_only>

:: ==========================================
:: 3. RUNTIME FILE SYSTEM INSPECTION (DELAYED EXPANSION & GEOMETRY)
:: ==========================================
setlocal enabledelayedexpansion
set LOG_COUNT=0

echo Scanning runtime log directory...
for /f "tokens=*" %%f in ('dir /b *.log 2^>nul') do (
    set /a LOG_COUNT+=1
    
    :: @ocalque:geometry:indent:pushd value="4"
    
    :: We access the runtime variable (!LOG_COUNT!) and compile-time metadata (TEAM_OWNER)
    :: @ocalque:replace_line --- echo [!LOG_COUNT!] Log %%f verified by SRE Team: {{ env.TEAM_OWNER }}
    echo [!LOG_COUNT!] Log %%f verified by SRE Team: Siemens-LocalDev-Team
    
    :: @ocalque:geometry:indent:popd
)

echo --- Setup Complete ---
