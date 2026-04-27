@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Devtools MCP - Distribution Build
echo ========================================
echo.
echo This script ensures the LATEST version is
echo built before distribution.
echo.

REM Check if Docker is available
docker --version >nul 2>&1
if errorlevel 1 (
  echo ERROR: Docker is not installed or not in PATH.
  echo Please install Docker Desktop from https://www.docker.com/products/docker-desktop
  exit /b 1
)

echo [1/3] Docker found: ✓
echo.

REM Create bin directory if it doesn't exist
if not exist bin mkdir bin
echo [2/3] Created bin directory (if needed): ✓
echo.

REM Get current timestamp for version tracking
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c%%a%%b)
for /f "tokens=1-2 delims=/:" %%a in ('time /t') do (set mytime=%%a%%b)

REM Build the executable (ALWAYS from fresh source)
echo [3/3] Building LATEST devtools-mcp.exe from source...
echo.
docker run --rm -v "%cd%":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/devtools-mcp.exe ."

if errorlevel 1 (
  echo.
  echo ERROR: Build failed. Check Docker is running and you have internet access.
  exit /b %errorlevel%
)

echo.
echo ========================================
echo ✓ Distribution Package Ready!
echo ========================================
echo.
echo Binary: bin\devtools-mcp.exe
echo Built:  %mydate% at %mytime%
echo.
echo Ready to distribute:
echo   - install.exe (for Windows users)
echo   - devtools-mcp.exe (for direct distribution)
echo.
echo Next steps:
echo   1. Run: build-installer.bat (to create install.exe)
echo   2. Distribute install.exe to end users
echo.
