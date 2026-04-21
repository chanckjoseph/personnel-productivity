@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Hello MCP Setup
echo ========================================
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

REM Build the executable
echo [3/3] Building hello-mcp.exe...
docker run --rm -v "%cd%":/src -w /src golang:1.22 sh -c "CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/hello-mcp.exe ."

if errorlevel 1 (
  echo.
  echo ERROR: Build failed. Check Docker is running and you have internet access.
  exit /b %errorlevel%
)

echo.
echo ========================================
echo ✓ Setup Complete!
echo ========================================
echo.
echo Binary built to: bin\hello-mcp.exe
echo.
echo Next steps:
echo   1. Restart VS Code
echo   2. The MCP will auto-connect via mcp.json config
echo   3. Try calling it via an AI tool
echo.
