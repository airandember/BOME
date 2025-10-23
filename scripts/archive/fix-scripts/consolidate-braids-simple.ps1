# Braid Consolidation Script
# Consolidates _backend and _frontend directories into unified _braids

$ErrorActionPreference = "Stop"

Write-Host "`nStarting Braid Consolidation..." -ForegroundColor Cyan

# Create log
$logFile = "braid-consolidation-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"

# Get list of braids
$braids = Get-ChildItem -Path "_backend" -Directory

Write-Host "Found $($braids.Count) braids to consolidate`n" -ForegroundColor Yellow

# Create _braids directory
if (Test-Path "_braids") {
    Write-Host "WARNING: _braids exists. Creating backup..." -ForegroundColor Yellow
    if (Test-Path "_braids.backup") {
        Remove-Item "_braids.backup" -Recurse -Force
    }
    Move-Item "_braids" "_braids.backup"
}

New-Item -ItemType Directory -Path "_braids" -Force | Out-Null

# Process each braid
$successCount = 0
foreach ($braid in $braids) {
    $braidName = $braid.Name
    Write-Host "Processing: $braidName" -ForegroundColor Cyan
    
    try {
        # Create braid directory
        $braidPath = "_braids/$braidName"
        New-Item -ItemType Directory -Path $braidPath -Force | Out-Null
        
        # Create frontend directory
        if (Test-Path "_frontend/$braidName") {
            Write-Host "  - Moving frontend content..." -ForegroundColor Gray
            New-Item -ItemType Directory -Path "$braidPath/frontend" -Force | Out-Null
            Get-ChildItem -Path "_frontend/$braidName" -Exclude "BRAID.md","README.md" | ForEach-Object {
                Copy-Item $_.FullName -Destination "$braidPath/frontend" -Recurse -Force
            }
        }
        
        # Create backend directory
        if (Test-Path "_backend/$braidName") {
            Write-Host "  - Moving backend content..." -ForegroundColor Gray
            New-Item -ItemType Directory -Path "$braidPath/backend" -Force | Out-Null
            Get-ChildItem -Path "_backend/$braidName" -Exclude "BRAID.md","README.md" | ForEach-Object {
                Copy-Item $_.FullName -Destination "$braidPath/backend" -Recurse -Force
            }
        }
        
        # Create unified BRAID.md
        Write-Host "  - Creating unified BRAID.md..." -ForegroundColor Gray
        $braidMd = "# Braid: $braidName`n`n"
        $braidMd += "**Architecture:** Full-Stack Braid (Frontend to Backend)`n"
        $braidMd += "**Last Updated:** $(Get-Date -Format 'yyyy-MM-dd')`n`n"
        $braidMd += "---`n`n"
        
        # Add backend content
        if (Test-Path "_backend/$braidName/BRAID.md") {
            $braidMd += "## Backend Architecture`n`n"
            $backendContent = Get-Content "_backend/$braidName/BRAID.md" -Raw
            $backendContent = $backendContent -replace '^\s*#[^#\n]*\n+', ''
            $braidMd += $backendContent
            $braidMd += "`n`n---`n`n"
        }
        
        # Add frontend content
        if (Test-Path "_frontend/$braidName/BRAID.md") {
            $braidMd += "## Frontend Architecture`n`n"
            $frontendContent = Get-Content "_frontend/$braidName/BRAID.md" -Raw
            $frontendContent = $frontendContent -replace '^\s*#[^#\n]*\n+', ''
            $braidMd += $frontendContent
            $braidMd += "`n`n---`n`n"
        }
        
        $braidMd += "## Integration Notes`n`n"
        $braidMd += "- Frontend: ``_braids/$braidName/frontend/```n"
        $braidMd += "- Backend: ``_braids/$braidName/backend/```n`n"
        $braidMd += "This braid represents a complete vertical slice of functionality.`n"
        
        $braidMd | Out-File -FilePath "$braidPath/BRAID.md" -Encoding UTF8
        
        Write-Host "  SUCCESS: $braidName" -ForegroundColor Green
        $successCount++
        
    } catch {
        Write-Host "  ERROR: $braidName - $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Create README
$readme = "# BOME Braids - Unified Architecture`n`n"
$readme += "**Last Updated:** $(Get-Date -Format 'yyyy-MM-dd')`n`n"
$readme += "## Overview`n`n"
$readme += "This directory contains the complete Braid Architecture for BOME.`n"
$readme += "Each braid is a full vertical slice from frontend to backend.`n`n"
$readme += "## Available Braids`n`n"
foreach ($braid in $braids) {
    $readme += "- **$($braid.Name)** - ``_braids/$($braid.Name)/```n"
}
$readme += "`n## Structure`n`n"
$readme += "Each braid contains:`n"
$readme += "- ``BRAID.md`` - Complete documentation`n"
$readme += "- ``frontend/`` - Frontend code and assets`n"
$readme += "- ``backend/`` - Backend code and assets`n`n"
$readme += "**Migration Date:** $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')`n"

$readme | Out-File -FilePath "_braids/README.md" -Encoding UTF8

# Summary
Write-Host "`n========================================" -ForegroundColor Green
Write-Host "CONSOLIDATION COMPLETE" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host "Braids Processed: $($braids.Count)" -ForegroundColor White
Write-Host "Successful: $successCount" -ForegroundColor Green
Write-Host "`nNew Structure: _braids/" -ForegroundColor Cyan
Write-Host "Documentation: _braids/README.md" -ForegroundColor Cyan
Write-Host "`nIMPORTANT: Original _backend and _frontend preserved for review.`n" -ForegroundColor Yellow

