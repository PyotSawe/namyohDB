@echo off
REM Build script for Relational Database project
REM Usage: scripts\build.bat [clean] [test] [release]

setlocal enabledelayedexpansion

if "%1"=="help" (
    echo Relational Database Build Script
    echo.
    echo Usage: scripts\build.bat [options]
    echo.
    echo Options:
    echo   clean    Clean build artifacts before building
    echo   test     Run tests after building
    echo   release  Build optimized release version
    echo   help     Show this help message
    echo.
    echo Examples:
    echo   scripts\build.bat                 # Basic build
    echo   scripts\build.bat clean test      # Clean, build, and test
    echo   scripts\build.bat release         # Release build
    goto :eof
)

echo Relational Database - Build Script
echo ===================================

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Go is not installed or not in PATH
    exit /b 1
)

REM Show Go version
echo.
echo Checking Go installation...
go version
echo   ^> Go installation verified

REM Clean if requested
for %%i in (%*) do (
    if "%%i"=="clean" (
        echo.
        echo Cleaning build artifacts...
        if exist bin rmdir /s /q bin
        if exist data rmdir /s /q data
        go clean -cache
        echo   ^> Clean completed
    )
)

REM Download dependencies
echo.
echo Downloading dependencies...
go mod download
if %errorlevel% neq 0 (
    echo Error: Failed to download dependencies
    exit /b 1
)
go mod tidy
echo   ^> Dependencies managed

REM Create bin directory
if not exist bin mkdir bin

REM Build the application
echo.
echo Building application...

set "BUILD_CMD=go build -o bin\relational-db.exe .\cmd\relational-db"

REM Check for release build
for %%i in (%*) do (
    if "%%i"=="release" (
        set "BUILD_CMD=go build -ldflags="-s -w" -o bin\relational-db.exe .\cmd\relational-db"
        echo   ^> Building optimized release version
    )
)

%BUILD_CMD%
if %errorlevel% neq 0 (
    echo Error: Build failed
    exit /b 1
)

echo   ^> Build completed successfully
echo   ^> Output: bin\relational-db.exe

REM Run tests if requested
for %%i in (%*) do (
    if "%%i"=="test" (
        echo.
        echo Running tests...
        echo   ^> Running unit tests...
        go test -v .\tests\unit\...
        if !errorlevel! neq 0 (
            echo Error: Unit tests failed
            exit /b 1
        )
        echo   ^> Running integration tests...
        go test -v .\tests\integration\...
        if !errorlevel! neq 0 (
            echo Error: Integration tests failed
            exit /b 1
        )
        echo   ^> All tests passed
    )
)

REM Format code
echo.
echo Formatting code...
go fmt .\...
echo   ^> Code formatted

REM Success message
echo.
echo Build completed successfully!
echo.
echo To run the database:
echo   bin\relational-db.exe
echo.
echo To run with environment variables:
echo   set DB_PORT=5433 ^& bin\relational-db.exe

endlocal