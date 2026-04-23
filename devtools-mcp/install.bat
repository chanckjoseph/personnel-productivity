@echo off
REM Devtools MCP One-File Installer (Batch + Binary)
REM This script can be embedded with devtools-mcp.exe to create a portable installer

setlocal enabledelayedexpansion

if "%~1"=="" (
  cls
  echo.
  echo ========================================
  echo Devtools MCP - Quick Install
  echo ========================================
  echo.
  echo This installer will set up Devtools MCP for use in your projects.
  echo.
  set /p INSTALL_PATH="Where would you like to install? (default: %%USERPROFILE%%\devtools-mcp): "
  
  if "!INSTALL_PATH!"=="" (
    set "INSTALL_PATH=%USERPROFILE%\devtools-mcp"
  )
) else (
  set "INSTALL_PATH=%~1"
)

echo.
echo Creating directory: !INSTALL_PATH!
if not exist "!INSTALL_PATH!" mkdir "!INSTALL_PATH!"

echo Copying binary...
copy "devtools-mcp.exe" "!INSTALL_PATH!\devtools-mcp.exe" >nul

if errorlevel 1 (
  echo ERROR: Failed to copy binary
  exit /b 1
)

echo Creating deployment guide...
(
  echo # Devtools MCP Installed Successfully
  echo.
  echo Installation Location: !INSTALL_PATH!
  echo.
  echo ## Next Steps - Configure in Your Project:
  echo.
  echo 1. In your project directory, create or edit `.vscode/mcp.json`:
  echo.
  echo ```json
  echo {
  echo   "mcpServers": {
  echo     "devtools-mcp": {
  echo       "type": "stdio",
  echo       "command": "!INSTALL_PATH!\devtools-mcp.exe"
  echo     }
  echo   }
  echo }
  echo ```
  echo.
  echo 2. Create credential files in your project root:
  echo    - `.pat` - GitHub Personal Access Token
  echo    - `.username` - GitHub username
  echo.
  echo 3. Restart VS Code
  echo.
) > "!INSTALL_PATH!\README.txt"

echo.
echo ========================================
echo ✓ Installation Complete!
echo ========================================
echo.
echo Binary installed to: !INSTALL_PATH!\devtools-mcp.exe
echo.
echo Use this path in your project's .vscode/mcp.json:
echo   "command": "!INSTALL_PATH!\devtools-mcp.exe"
echo.
echo For more details, see: !INSTALL_PATH!\README.txt
echo.
pause

