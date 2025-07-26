#!/usr/bin/env powershell

# Run script for GoVaultFS
Write-Host "Running GoVaultFS..." -ForegroundColor Green

# First build the application
& .\build.ps1

if ($LASTEXITCODE -eq 0) {
    # Run the built executable
    & .\bin\fs.exe
} else {
    Write-Host "Cannot run - build failed!" -ForegroundColor Red
    exit $LASTEXITCODE
}
