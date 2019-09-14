@echo off
for %%i in (.) do set folder=%%~nxi
../build.bat %folder%
