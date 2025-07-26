#!/usr/bin/env powershell

# Build script for GoVaultFS
Write-Host "Building GoVaultFS..." -ForegroundColor Green

# Create bin directory if it doesn't exist
if (!(Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin"
}

# Build the Go application
go build -o bin/fs.exe

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build completed successfully!" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit $LASTEXITCODE
}
