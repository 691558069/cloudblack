@echo off
setlocal
cd /d "%~dp0"

echo CloudBlack Windows Launcher
echo.
set /p PORT=Please enter port [8080]: 
if "%PORT%"=="" set PORT=8080

echo.
echo Starting CloudBlack on port %PORT% ...
echo Open: http://127.0.0.1:%PORT%/
echo.
cloudblack.exe -port %PORT%

echo.
echo CloudBlack has stopped.
pause
