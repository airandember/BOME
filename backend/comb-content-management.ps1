# Comb All Content Management Strands

Write-Host ""
Write-Host "CONTENT MANAGEMENT BRAID - COMPLETE COMBING" -ForegroundColor DarkYellow
Write-Host "============================================" -ForegroundColor DarkYellow
Write-Host ""

# STRAND 1: Content Creation and Publishing
Write-Host "STRAND 1: Content Creation and Publishing" -ForegroundColor Yellow
Write-Host "------------------------------------------" -ForegroundColor Yellow

$hasContent = Select-String -Path "content/**/*.go" -Pattern "func.*Create|func.*Publish|func.*Article" -Quiet 2>$null
if ($hasContent) { 
    Write-Host "  [OK] Content functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Content functions missing" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 2: Category Management
Write-Host "STRAND 2: Category Management" -ForegroundColor Yellow
Write-Host "-----------------------------" -ForegroundColor Yellow

$hasCategories = Select-String -Path "**/*.go" -Pattern "func.*Category|func.*Categor" -Quiet 2>$null
if ($hasCategories) { 
    Write-Host "  [OK] Category functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Categories missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 3: Tag System
Write-Host "STRAND 3: Tag System" -ForegroundColor Yellow
Write-Host "--------------------" -ForegroundColor Yellow

if (Test-Path "content/models/tags.go") {
    Write-Host "  [OK] tags.go model exists" -ForegroundColor Green
    
    $hasTagFuncs = Select-String -Path "content/models/tags.go" -Pattern "func.*" | Measure-Object
    Write-Host "  [INFO] Found $($hasTagFuncs.Count) functions in tags.go" -ForegroundColor Cyan
} else {
    Write-Host "  [SPLIT-END] tags.go model missing" -ForegroundColor Red
}

Write-Host ""

# STRAND 4: Content Search
Write-Host "STRAND 4: Content Search and Discovery" -ForegroundColor Yellow
Write-Host "---------------------------------------" -ForegroundColor Yellow

$hasSearch = Select-String -Path "**/*.go" -Pattern "func.*Search|func.*Find" -Quiet 2>$null
if ($hasSearch) { 
    Write-Host "  [OK] Search functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Search missing" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 5: SEO and Metadata
Write-Host "STRAND 5: SEO and Metadata" -ForegroundColor Yellow
Write-Host "--------------------------" -ForegroundColor Yellow

$hasSEO = Select-String -Path "**/*.go" -Pattern "SEO|Metadata|MetaDescription" -Quiet 2>$null
if ($hasSEO) { 
    Write-Host "  [OK] SEO/Metadata found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] SEO/Metadata missing" -ForegroundColor Yellow 
}

Write-Host ""

# Check content package structure
Write-Host "CONTENT PACKAGE STRUCTURE:" -ForegroundColor Cyan

if (Test-Path "content") {
    Write-Host "  [OK] content/ directory exists" -ForegroundColor Green
    
    if (Test-Path "content/models") {
        Write-Host "  [OK] content/models/ exists" -ForegroundColor Green
        $modelCount = (Get-ChildItem "content/models/*.go" -File 2>$null | Measure-Object).Count
        Write-Host "  [INFO] Found $modelCount model files" -ForegroundColor Cyan
    }
    
    if (Test-Path "content/handlers") {
        Write-Host "  [OK] content/handlers/ exists" -ForegroundColor Green
        $handlerCount = (Get-ChildItem "content/handlers/*.go" -File 2>$null | Measure-Object).Count
        Write-Host "  [INFO] Found $handlerCount handler files" -ForegroundColor Cyan
    }
} else {
    Write-Host "  [WARNING] No dedicated content/ directory" -ForegroundColor Yellow
    Write-Host "  [INFO] Content might be in admin or other packages" -ForegroundColor Gray
}

Write-Host ""
Write-Host "ALL 5 STRANDS COMBED" -ForegroundColor Green
Write-Host ""

