TITLE Development Tools for Abi

:: BUILD VARIABLES
set BUILD_MCTOKEN=true
set BUILD_TOWEL=true

:: CREATE BIN
mkdir c:\bin
echo off
SETX PATH "c:\bin;%PATH%"
echo on

:: ADD ALL BIN FOLDERS

mkdir c:\bin\mctoken

echo Building MCTOKEN
echo off
:: Building Mctoken
if "%BUILD_MCTOKEN%"=="true" (
    cd %~dp0minecraft\mctoken
    go build .
    move %~dp0\minecraft\mctoken\mctoken.exe "C:\bin"
    cd ..\..
)
echo on
Building Towel
echo off
:: Building Towel
if "%BUILD_TOWEL%"=="true" (
    cd %~dp0minecraft\towel
    go build .
    move %~dp0\minecraft\towel\towel.exe "C:\bin"
    cd ..\..
)
echo on
pause
