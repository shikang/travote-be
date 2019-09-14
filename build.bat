@echo off
pushd %~dp0

if "%1"=="" (
	echo "Please specify build folder."
	echo "build.bat <lambda_build_folder>"
	pause
	exit 1
)

if exist out (
	RD /S /Q out
)

mkdir out

echo "Copying Dependencies..."
copy /Y utils\*.go out\.
copy /Y common\*.go out\.
copy /Y %1\*.go out\.

echo "Building %1 ..."
cd out

setlocal enabledelayedexpansion

echo "Setting GO OS to Linux..."
set GOOS=linux

set gofiles=
for %%i in (*.go) do set "gofiles=!gofiles! %%i"

echo "Building..."
go build -o main %gofiles%

echo "Zipping..."
%GOPATH%\bin\build-lambda-zip.exe -o main.zip main

echo "Output: ../build/main.zip"

popd
pause
