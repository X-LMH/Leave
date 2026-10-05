@echo off
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0connect-phone.ps1" %*
set "phoneConnectExit=%errorlevel%"
echo.
pause
exit /b %phoneConnectExit%
