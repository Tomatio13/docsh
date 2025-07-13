@echo off
REM Go-Zsh Windows 64bit Build Script

echo Building Go-Zsh for Windows 64bit...

REM 環境変数設定
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

REM ビルド実行
go build -ldflags="-s -w" -o go-zsh-windows-x64.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful: go-zsh-windows-x64.exe
    echo File size:
    dir go-zsh-windows-x64.exe | findstr "go-zsh-windows-x64.exe"
) else (
    echo Build failed with error code %ERRORLEVEL%
    exit /b %ERRORLEVEL%
)

REM 32bit版も作成（互換性のため）
echo.
echo Building Go-Zsh for Windows 32bit...
set GOARCH=386
go build -ldflags="-s -w" -o go-zsh-windows-x86.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful: go-zsh-windows-x86.exe
    echo File size:
    dir go-zsh-windows-x86.exe | findstr "go-zsh-windows-x86.exe"
) else (
    echo 32bit build failed with error code %ERRORLEVEL%
)

echo.
echo Build completed!