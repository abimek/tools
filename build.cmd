TITLE Development Tools for Abi

:: Building Towel
rem Building Towel...
cd %~dp0minecraft\towel
go build towel.go decrypt.go
mkdir c:\bin
echo off
SETX PATH "c:\bin;%PATH%"
echo on
move %CD%\towel.exe "C:\bin"
rem Towel Built...
cd ..\..


pause
