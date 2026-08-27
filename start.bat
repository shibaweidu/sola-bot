@echo off
setlocal
cd /d "%~dp0"
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\start.ps1" %*
set "exitCode=%ERRORLEVEL%"
if not "%exitCode%"=="0" (
    echo.
    echo Sola startup failed with exit code %exitCode%.
)
pause
exit /b %exitCode%
