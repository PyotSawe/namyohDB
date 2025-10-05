# Build script for Relational Database project
# Usage: .\scripts\build.ps1 [clean] [test] [release]

param(
    [switch]$Clean,
    [switch]$Test,
    [switch]$Release,
    [switch]$Help
)

if ($Help) {
    Write-Host "Relational Database Build Script"
    Write-Host ""
    Write-Host "Usage: .\scripts\build.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -Clean    Clean build artifacts before building"
    Write-Host "  -Test     Run tests after building"
    Write-Host "  -Release  Build optimized release version"
    Write-Host "  -Help     Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\scripts\build.ps1                    # Basic build"
    Write-Host "  .\scripts\build.ps1 -Clean -Test       # Clean, build, and test"
    Write-Host "  .\scripts\build.ps1 -Release           # Release build"
    exit 0
}

Write-Host "Relational Database - Build Script" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan

# Get the project root directory
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Definition)
Push-Location $ProjectRoot

try {
    # Clean if requested
    if ($Clean) {
        Write-Host ""
        Write-Host "Cleaning build artifacts..." -ForegroundColor Yellow
        
        if (Test-Path "bin") {
            Remove-Item -Recurse -Force "bin"
            Write-Host "  ✓ Removed bin directory"
        }
        
        if (Test-Path "data") {
            Remove-Item -Recurse -Force "data"
            Write-Host "  ✓ Removed data directory"
        }
        
        # Clean Go module cache
        go clean -cache
        Write-Host "  ✓ Cleaned Go cache"
    }
    
    # Verify Go installation
    Write-Host ""
    Write-Host "Checking Go installation..." -ForegroundColor Yellow
    $goVersion = go version
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Go is not installed or not in PATH"
        exit 1
    }
    Write-Host "  ✓ $goVersion"
    
    # Download dependencies
    Write-Host ""
    Write-Host "Downloading dependencies..." -ForegroundColor Yellow
    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to download dependencies"
        exit 1
    }
    Write-Host "  ✓ Dependencies downloaded"
    
    # Tidy modules
    go mod tidy
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to tidy modules"
        exit 1
    }
    Write-Host "  ✓ Modules tidied"
    
    # Create bin directory
    if (!(Test-Path "bin")) {
        New-Item -ItemType Directory -Name "bin" | Out-Null
    }
    
    # Build the application
    Write-Host ""
    Write-Host "Building application..." -ForegroundColor Yellow
    
    $buildCmd = "go build"
    $outputPath = "bin/relational-db.exe"
    
    if ($Release) {
        # Release build with optimizations
        $buildCmd += " -ldflags=""-s -w"""
        Write-Host "  • Building optimized release version"
    } else {
        Write-Host "  • Building debug version"
    }
    
    $buildCmd += " -o $outputPath ./cmd/relational-db"
    
    Invoke-Expression $buildCmd
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed"
        exit 1
    }
    
    Write-Host "  ✓ Build completed successfully"
    
    # Get build information
    $buildInfo = Get-Item $outputPath
    $buildSize = [math]::Round($buildInfo.Length / 1KB, 2)
    Write-Host "  • Output: $outputPath ($buildSize KB)"
    
    # Run tests if requested
    if ($Test) {
        Write-Host ""
        Write-Host "Running tests..." -ForegroundColor Yellow
        
        # Run unit tests
        Write-Host "  • Running unit tests..."
        go test ./tests/unit/... -v
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Unit tests failed"
            exit 1
        }
        
        # Run integration tests
        Write-Host "  • Running integration tests..."
        go test ./tests/integration/... -v
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Integration tests failed"
            exit 1
        }
        
        Write-Host "  ✓ All tests passed"
    }
    
    # Format code
    Write-Host ""
    Write-Host "Formatting code..." -ForegroundColor Yellow
    go fmt ./...
    Write-Host "  ✓ Code formatted"
    
    # Final success message
    Write-Host ""
    Write-Host "Build completed successfully! 🎉" -ForegroundColor Green
    Write-Host ""
    Write-Host "To run the database:" -ForegroundColor Cyan
    Write-Host "  .\bin\relational-db.exe" -ForegroundColor White
    Write-Host ""
    Write-Host "To run with environment variables:" -ForegroundColor Cyan
    Write-Host "  `$env:DB_PORT=5433; .\bin\relational-db.exe" -ForegroundColor White
    
} finally {
    Pop-Location
}