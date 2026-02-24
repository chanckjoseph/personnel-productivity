@echo off
REM This script is designed to be run by Windows Task Scheduler at logon or startup.

REM Change directory to the folder containing this script
cd /d "%~dp0"

echo --- Starting Personnel Productivity Service ---
python manage_docker.py up

REM Optional: Keep the window open for a few seconds to see status
timeout /t 5
