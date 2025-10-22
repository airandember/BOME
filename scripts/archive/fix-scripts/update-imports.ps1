# 🔄 Backend Import Update Script
# This script updates all Go imports from old structure to new braid structure

Write-Host "🚀 Starting Backend Import Updates..." -ForegroundColor Cyan
Write-Host ""

$replacements = @(
    # Package declarations
    @{ Old = 'package routes'; New = 'package handlers' },
    @{ Old = 'package database'; New = 'package models' },
    
    # Database type references
    @{ Old = '*database.DB'; New = '*models.DB' },
    @{ Old = 'database.AuditLog'; New = 'models.AuditLog' },
    @{ Old = 'database.Session'; New = 'models.Session' },
    @{ Old = 'database.User'; New = 'models.User' },
    @{ Old = 'database.Video'; New = 'models.Video' },
    @{ Old = 'database.Subscription'; New = 'models.Subscription' },
    @{ Old = 'database.SubscriptionPlan'; New = 'models.SubscriptionPlan' },
    @{ Old = 'database.StripeEntity'; New = 'models.StripeEntity' },
    @{ Old = 'database.Advertisement'; New = 'models.Advertisement' },
    @{ Old = 'database.Tag'; New = 'models.Tag' },
    @{ Old = 'database.Analytics'; New = 'models.Analytics' },
    
    # Import path updates - Authentication
    @{ Old = '"bome-backend/internal/database"'; New = '"bome-backend/authentication/models"' },
    @{ Old = '"bome-backend/internal/services"'; New = '"bome-backend/authentication/services"' },
    @{ Old = '"bome-backend/internal/middleware"'; New = '"bome-backend/authentication/middleware"' },
    @{ Old = '"bome-backend/internal/routes"'; New = '"bome-backend/authentication/handlers"' },
    
    # Infrastructure imports
    @{ Old = '"bome-backend/internal/cache"'; New = '"bome-backend/infrastructure/cache"' },
    @{ Old = '"bome-backend/internal/config"'; New = '"bome-backend/infrastructure/config"' }
)

# Get all .go files in backend/ (excluding go.mod, go.sum)
$goFiles = Get-ChildItem -Path "backend" -Recurse -Filter "*.go" | Where-Object { $_.Name -ne "go.mod" -and $_.Name -ne "go.sum" }

Write-Host "Found $($goFiles.Count) Go files to process" -ForegroundColor Yellow
Write-Host ""

$updatedCount = 0

foreach ($file in $goFiles) {
    Write-Host "Processing: $($file.FullName.Replace($PWD, '.'))" -ForegroundColor Gray
    
    $content = Get-Content $file.FullName -Raw
    $originalContent = $content
    
    # Apply all replacements
    foreach ($replacement in $replacements) {
        $content = $content -replace [regex]::Escape($replacement.Old), $replacement.New
    }
    
    # If content changed, write it back
    if ($content -ne $originalContent) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
        $updatedCount++
        Write-Host "  ✅ Updated" -ForegroundColor Green
    } else {
        Write-Host "  - No changes needed" -ForegroundColor DarkGray
    }
}

Write-Host ""
Write-Host "🎊 Import Update Complete!" -ForegroundColor Green
Write-Host "  Files processed: $($goFiles.Count)" -ForegroundColor Cyan
Write-Host "  Files updated: $updatedCount" -ForegroundColor Cyan
Write-Host ""
Write-Host "⚠️ Note: Some files may need braid-specific imports (video-streaming, subscription, etc.)"
Write-Host "   These will need manual review and updating."
Write-Host ""
Write-Host "Next step: Run 'go build' in backend/ to check for errors"

