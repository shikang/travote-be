@echo off
pushd %~dp0

if "%1"=="" (
	echo "Please specify build folder."
	echo "build.bat <lambda_build_folder>"
	pause
	exit 1
)

if exist build (
	RD /S /Q build
)

mkdir build

echo "Copying Dependencies..."
copy /Y utils\*.go build\.
copy /Y structs\*.go build\.
copy /Y %1\*.go build\.

if exist %1\dependencies.txt (
	for /F "tokens=*" %%i in (%1\dependencies.txt) do (
		echo "Adding %%i Dependencies..."
		copy /Y %%i\*.go build\.
	)
) else (
	echo "No additional Dependencies!"
)

echo "Building %1 ..."
cd build

setlocal enabledelayedexpansion

echo "Setting GO OS to Linux..."
set GOOS=linux

set gofiles=
for %%i in (*.go) do set "gofiles=!gofiles! %%i"

echo "Building..."
go build -o main %gofiles%

echo "Zipping..."
%GOPATH%\bin\build-lambda-zip.exe -o main.zip main

cd ..

if exist out (
	RD /S /Q out
)

mkdir out

copy /Y build\main.zip out\.

echo "Output: ../out/main.zip"

popd
pause
