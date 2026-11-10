@echo off
setlocal
cd /d "%~dp0"
echo.
echo  [ CYBERCIRCUIT GOTHAM 0.5.0 ]  Windows build
echo.
where go >nul 2>nul
if errorlevel 1 (
  echo Go was not found.
  echo Install Go 1.23 or newer from https://go.dev/dl/ and run this file again.
  pause
  exit /b 1
)
for /f "tokens=3" %%V in ('go version') do echo Detected %%V
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -trimpath -ldflags="-s -w -H=windowsgui" -o CyberCircuitGotham.exe .
if errorlevel 1 (
  echo.
  echo Build failed. Check the error above.
  pause
  exit /b 1
)
echo.
echo Build successful: %CD%\CyberCircuitGotham.exe
pause
