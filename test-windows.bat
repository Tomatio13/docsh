@echo off
REM Cherry Shell Windows 10/11 64-bit テストスクリプト
echo 🌸 Cherry Shell Windows Test 🌸
echo.
echo Testing Cherry Shell on Windows 10/11 64-bit...
echo.

REM 実行ファイルの存在確認
if not exist "cherrysh-windows-x64.exe" (
    echo ERROR: cherrysh-windows-x64.exe not found!
    echo Please make sure the file is in the current directory.
    pause
    exit /b 1
)

echo ✓ Cherry Shell executable found
echo.

REM ファイルサイズ確認
for %%A in (cherrysh-windows-x64.exe) do echo File size: %%~zA bytes
echo.

REM 基本動作テスト
echo Testing basic functionality...
echo ls > test_input.txt
echo pwd >> test_input.txt
echo exit >> test_input.txt

echo Running Cherry Shell with test commands...
echo.
cherrysh-windows-x64.exe < test_input.txt

REM クリーンアップ
del test_input.txt 2>nul

echo.
echo 🌸 Cherry Shell Windows Test Complete! 🌸
pause