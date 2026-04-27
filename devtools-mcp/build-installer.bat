@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Devtools MCP - Build Windows Installer
echo ========================================
echo.

REM Check if binary exists
if not exist bin\devtools-mcp.exe (
  echo ERROR: bin\devtools-mcp.exe not found.
  echo.
  echo Please run: distribute.bat
  echo to build the binary first.
  exit /b 1
)

echo [1/2] Found binary: bin\devtools-mcp.exe ✓
echo.

REM Check for InnoSetup
echo [2/2] Checking for InnoSetup compiler...
set INNO_PATH=C:\Program Files (x86)\Inno Setup 6
if not exist "%INNO_PATH%\ISCC.exe" (
  set INNO_PATH=C:\Program Files\Inno Setup 6
)

if not exist "%INNO_PATH%\ISCC.exe" (
  echo.
  echo WARNING: InnoSetup 6 not found in default location.
  echo.
  echo To create install.exe, you need to:
  echo   1. Download InnoSetup 6: https://jrsoftware.org/isdl.php
  echo   2. Run the installer.iss manually from:
  echo      Right-click installer.iss ^> Compile with InnoSetup
  echo.
  echo OR use the command:
  echo   "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" installer.iss
  echo.
) else (
  echo InnoSetup found: ✓
  echo.
  echo Building installer...
  "%INNO_PATH%\ISCC.exe" installer.iss
  
  if errorlevel 1 (
    echo.
    echo ERROR: Installer build failed.
    exit /b %errorlevel%
  )
  echo.
  echo ========================================
  echo ✓ Install.exe created successfully!
  echo ========================================
  echo.
  echo Location: install.exe
  echo.
  echo This installer can now be distributed to other projects.
  echo Users will run it to install devtools-mcp globally and then
  echo configure it in their project's .vscode/mcp.json
  echo.
)

