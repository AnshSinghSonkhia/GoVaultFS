#!/usr/bin/env powershell

# Setup script to add GnuWin32 Make to PATH permanently
Write-Host "Setting up Make for Windows..." -ForegroundColor Green

$gnuWinPath = "C:\Program Files (x86)\GnuWin32\bin"

# Check if GnuWin32 is installed
if (Test-Path $gnuWinPath) {
    # Get current user PATH
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    
    # Check if GnuWin32 is already in PATH
    if ($userPath -notlike "*$gnuWinPath*") {
        # Add to user PATH
        $newPath = $userPath + ";" + $gnuWinPath
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Host "Added GnuWin32 Make to your user PATH permanently." -ForegroundColor Green
        Write-Host "Please restart your terminal or VS Code for the changes to take effect." -ForegroundColor Yellow
    } else {
        Write-Host "GnuWin32 Make is already in your PATH." -ForegroundColor Green
    }
    
    # Also add to current session
    $env:PATH += ";$gnuWinPath"
    Write-Host "Added to current session PATH as well." -ForegroundColor Green
    
} else {
    Write-Host "GnuWin32 Make not found. Please install it first using:" -ForegroundColor Red
    Write-Host "winget install GnuWin32.Make" -ForegroundColor Cyan
}
