# ============================================================================
# 🧬 BRAID CONSOLIDATION SCRIPT
# ============================================================================
# Consolidates _backend and _frontend directories into unified _braids
# ============================================================================

$ErrorActionPreference = "Stop"

Write-Host "`n╔════════════════════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "║               🧬 BRAID ARCHITECTURE CONSOLIDATION                              ║" -ForegroundColor Cyan
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "║     Unifying _backend and _frontend into _braids for complete context         ║" -ForegroundColor White
Write-Host "║                                                                                ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════════════════════╝`n" -ForegroundColor Cyan

# Create log
$logFile = "braid-consolidation-$(Get-Date -Format 'yyyyMMdd-HHmmss').log"
function Log($message) {
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    "$timestamp - $message" | Out-File -Append -FilePath $logFile
    Write-Host $message
}

Log "🚀 Starting Braid Consolidation..."

# Get list of braids from _backend
$braids = Get-ChildItem -Path "_backend" -Directory | Where-Object { $_.Name -notlike ".*" }

Log "`n📋 Found $($braids.Count) braids to consolidate:"
foreach ($braid in $braids) {
    Log "   • $($braid.Name)"
}

# Create _braids directory
if (Test-Path "_braids") {
    Log "`n⚠️  _braids directory already exists. Backing up to _braids.backup..."
    if (Test-Path "_braids.backup") {
        Remove-Item "_braids.backup" -Recurse -Force
    }
    Move-Item "_braids" "_braids.backup"
}

Log "`n📁 Creating _braids directory structure..."
New-Item -ItemType Directory -Path "_braids" -Force | Out-Null

# Process each braid
$successCount = 0
$errorCount = 0

foreach ($braid in $braids) {
    $braidName = $braid.Name
    Log "`n🔄 Processing: $braidName"
    
    try {
        # Create braid directory
        $braidPath = "_braids/$braidName"
        New-Item -ItemType Directory -Path $braidPath -Force | Out-Null
        
        # Create frontend directory if frontend exists
        if (Test-Path "_frontend/$braidName") {
            Log "   ✅ Moving frontend content..."
            New-Item -ItemType Directory -Path "$braidPath/frontend" -Force | Out-Null
            
            # Copy all frontend content except BRAID.md
            Get-ChildItem -Path "_frontend/$braidName" -Exclude "BRAID.md","README.md" | ForEach-Object {
                Copy-Item $_.FullName -Destination "$braidPath/frontend" -Recurse -Force
            }
        } else {
            Log "   ⚠️  No frontend directory found"
        }
        
        # Create backend directory if backend exists
        if (Test-Path "_backend/$braidName") {
            Log "   ✅ Moving backend content..."
            New-Item -ItemType Directory -Path "$braidPath/backend" -Force | Out-Null
            
            # Copy all backend content except BRAID.md and README.md
            Get-ChildItem -Path "_backend/$braidName" -Exclude "BRAID.md","README.md" | ForEach-Object {
                Copy-Item $_.FullName -Destination "$braidPath/backend" -Recurse -Force
            }
        } else {
            Log "   ⚠️  No backend directory found"
        }
        
        # Merge BRAID.md files
        Log "   📝 Creating unified BRAID.md..."
        $unifiedBraid = @()
        
        $unifiedBraid += "# 🧬 $braidName Braid - Complete Vertical Slice"
        $unifiedBraid += ""
        $unifiedBraid += "**Architecture:** Full-Stack Braid (Frontend → Backend)"
        $unifiedBraid += "**Last Updated:** $(Get-Date -Format 'yyyy-MM-dd')"
        $unifiedBraid += "**Status:** Consolidated"
        $unifiedBraid += ""
        $unifiedBraid += "---"
        $unifiedBraid += ""
        $unifiedBraid += "## 📋 Table of Contents"
        $unifiedBraid += ""
        $unifiedBraid += "1. [Overview](#overview)"
        $unifiedBraid += "2. [Frontend Architecture](#frontend-architecture)"
        $unifiedBraid += "3. [Backend Architecture](#backend-architecture)"
        $unifiedBraid += "4. [Data Flow](#data-flow)"
        $unifiedBraid += "5. [API Contract](#api-contract)"
        $unifiedBraid += ""
        $unifiedBraid += "---"
        $unifiedBraid += ""
        
        # Add backend BRAID.md content if exists
        if (Test-Path "_backend/$braidName/BRAID.md") {
            $unifiedBraid += "## 🔙 Backend Architecture"
            $unifiedBraid += ""
            $unifiedBraid += "**Source:** \`_backend/$braidName/BRAID.md\`"
            $unifiedBraid += ""
            $backendContent = Get-Content "_backend/$braidName/BRAID.md" -Raw
            # Remove the first heading line if it exists
            $backendContent = $backendContent -replace '^#[^#\n]*\n+', ''
            $unifiedBraid += $backendContent
            $unifiedBraid += ""
            $unifiedBraid += "---"
            $unifiedBraid += ""
        }
        
        # Add frontend BRAID.md content if exists
        if (Test-Path "_frontend/$braidName/BRAID.md") {
            $unifiedBraid += "## 🎨 Frontend Architecture"
            $unifiedBraid += ""
            $unifiedBraid += "**Source:** \`_frontend/$braidName/BRAID.md\`"
            $unifiedBraid += ""
            $frontendContent = Get-Content "_frontend/$braidName/BRAID.md" -Raw
            # Remove the first heading line if it exists
            $frontendContent = $frontendContent -replace '^#[^#\n]*\n+', ''
            $unifiedBraid += $frontendContent
            $unifiedBraid += ""
            $unifiedBraid += "---"
            $unifiedBraid += ""
        }
        
        $unifiedBraid += "## 🔗 Integration Notes"
        $unifiedBraid += ""
        $unifiedBraid += "**Frontend Location:** \`_braids/$braidName/frontend/\`"
        $unifiedBraid += "**Backend Location:** \`_braids/$braidName/backend/\`"
        $unifiedBraid += ""
        $unifiedBraid += "This braid represents a complete vertical slice of functionality from the user interface down to the data persistence layer."
        $unifiedBraid += ""
        
        # Write unified BRAID.md
        $unifiedBraid | Out-File -FilePath "$braidPath/BRAID.md" -Encoding UTF8
        
        Log "   ✅ Successfully consolidated $braidName"
        $successCount++
        
    } catch {
        Log "   ❌ ERROR: $($_.Exception.Message)"
        $errorCount++
    }
}

# Create README for _braids directory
Log "`n📄 Creating _braids README..."
$braidsReadme = @"
# 🧬 BOME Braids - Unified Architecture

**Last Updated:** $(Get-Date -Format 'yyyy-MM-dd')

## 📋 Overview

This directory contains the complete **Braid Architecture** for BOME. Each braid represents a full vertical slice of functionality from frontend to backend.

## 🏗️ Structure

``````
_braids/
  <braid-name>/
    BRAID.md              # Complete documentation (frontend + backend)
    frontend/             # All frontend code and assets
      layers/
        presentation/     # UI components, pages
        state-management/ # State, stores, hooks
    backend/              # All backend code and assets
      layers/
        application/      # API routes, controllers
        business-logic/   # Services, use cases
        data-access/      # Repositories, DAOs
        persistence/      # Database schemas, migrations
``````

## 🧬 Available Braids

$(($braids | ForEach-Object { "- **$($_.Name)** - \`_braids/$($_.Name)/\`" }) -join "`n")

## 🎯 Philosophy

Each braid is a **complete, self-contained feature** that can be understood and developed independently. The braid structure ensures:

1. ✅ **Complete Context** - Frontend and backend together
2. ✅ **Vertical Slicing** - End-to-end functionality in one place
3. ✅ **Clear Boundaries** - Well-defined interfaces between braids
4. ✅ **Easy Navigation** - All related code in one directory

## 📚 Documentation

Each braid has a unified \`BRAID.md\` file that contains:
- Complete architecture overview
- Frontend layer documentation
- Backend layer documentation
- Data flow diagrams
- API contracts
- Integration notes

## 🔄 Migration

This structure was consolidated from:
- \`_backend/<braid-name>/\` → \`_braids/<braid-name>/backend/\`
- \`_frontend/<braid-name>/\` → \`_braids/<braid-name>/frontend/\`

**Migration Date:** $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
**Migration Log:** \`$logFile\`

## 🚀 Next Steps

1. Update code references to use new paths
2. Update documentation links
3. Configure IDE/editor to recognize new structure
4. Update deployment scripts if needed

---

**Note:** The original \`_backend\` and \`_frontend\` directories have been preserved for reference and can be removed once migration is verified.
"@

$braidsReadme | Out-File -FilePath "_braids/README.md" -Encoding UTF8

# Summary
Log "`n╔════════════════════════════════════════════════════════════════════════════════╗"
Log "║                        CONSOLIDATION COMPLETE                                  ║"
Log "╚════════════════════════════════════════════════════════════════════════════════╝`n"
Log "📊 Summary:"
Log "   • Braids Processed: $($braids.Count)"
Log "   • Successful: $successCount ✅"
Log "   • Errors: $errorCount ❌"
Log "   • Log File: $logFile"
Log "`n📁 New Structure: _braids/"
Log "📄 Documentation: _braids/README.md"
Log "`n⚠️  IMPORTANT: The original _backend and _frontend directories are preserved."
Log "   Review the new _braids structure before deleting them.`n"

# List the new structure
Log "🔍 New _braids structure:"
Get-ChildItem "_braids" -Directory | ForEach-Object {
    Log "   📦 $($_.Name)/"
    if (Test-Path "$($_.FullName)/frontend") {
        Log "      └─ frontend/"
    }
    if (Test-Path "$($_.FullName)/backend") {
        Log "      └─ backend/"
    }
}

Log "`n✅ Consolidation complete! Check $logFile for details.`n"

