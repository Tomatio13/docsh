@echo off
REM CherryShell Windows Build Script

echo Building CherryShell for Windows...

REM Create Release directory if it doesn't exist
if not exist "Release" mkdir Release

REM 環境変数設定
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

REM ビルド実行 (64bit)
echo Building for Windows 64bit...
go build -ldflags="-s -w" -o Release\cherrysh-windows-x64.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful: Release\cherrysh-windows-x64.exe
    echo File size:
    dir Release\cherrysh-windows-x64.exe | findstr "cherrysh-windows-x64.exe"
) else (
    echo Build failed with error code %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)

REM 32bit版も作成（互換性のため）
echo.
echo Building for Windows 32bit...
set GOARCH=386
go build -ldflags="-s -w" -o Release\cherrysh-windows-x86.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful: Release\cherrysh-windows-x86.exe
    echo File size:
    dir Release\cherrysh-windows-x86.exe | findstr "cherrysh-windows-x86.exe"
) else (
    echo 32bit build failed with error code %ERRORLEVEL%
)

echo.
echo Build completed! Available binaries in Release folder:
echo Windows 64bit: Release\cherrysh-windows-x64.exe
echo Windows 32bit: Release\cherrysh-windows-x86.exe