#!/usr/bin/env powershell

# Test script for GoVaultFS
Write-Host "Running tests for GoVaultFS..." -ForegroundColor Green

# Run Go tests with verbose output
go test ./... -v

if ($LASTEXITCODE -eq 0) {
    Write-Host "All tests passed!" -ForegroundColor Green
} else {
    Write-Host "Some tests failed!" -ForegroundColor Red
    exit $LASTEXITCODE
}
