@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Devtools MCP Setup + Installer Build
echo ========================================
echo.

REM Check if Docker is available
docker --version >nul 2>&1
if errorlevel 1 (
  echo ERROR: Docker is not installed or not in PATH.
  echo Please install Docker Desktop from https://www.docker.com/products/docker-desktop
  exit /b 1
)

echo [1/4] Docker found: ✓
echo.

REM Create bin directory if it doesn't exist
if not exist bin mkdir bin
echo [2/4] Created bin directory (if needed): ✓
echo.

REM Build the Windows executable
echo [3/4] Building devtools-mcp.exe...
docker run --rm -v "%cd%":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/devtools-mcp.exe ."

if errorlevel 1 (
  echo.
  echo ERROR: Build failed. Check Docker is running and you have internet access.
  exit /b %errorlevel%
)

echo devtools-mcp.exe built successfully: ✓
echo.

REM Check for InnoSetup
echo [4/4] Checking for InnoSetup compiler...
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

echo.
echo Binary built to: bin\devtools-mcp.exe
echo.

