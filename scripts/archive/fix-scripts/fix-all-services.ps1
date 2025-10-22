# Fix all service files to use database.DB and add proper imports

Write-Host "🔄 Fixing all service files..."

$serviceDirs = @(
    "backend/video-streaming/services",
    "backend/analytics/services",
    "backend/authentication/services",
    "backend/communication/services",
    "backend/subscription/services"
)

foreach ($dir in $serviceDirs) {
    if (Test-Path $dir) {
        $files = Get-ChildItem -Path $dir -Filter "*.go"
        foreach ($file in $files) {
            Write-Host "  Processing: $($file.Name)" -ForegroundColor Gray
            
            $content = Get-Content $file.FullName -Raw
            $originalContent = $content
            
            # Add database import if not present and file uses models.DB or database.DB
            if ($content -match 'models\.DB|database\.DB' -and $content -notmatch '"bome-backend/infrastructure/database"') {
                $content = $content -replace '(import \()', '$1`n`t"bome-backend/infrastructure/database"'
            }
            
            # Replace models.DB with database.DB
            $content = $content -replace '\*models\.DB', '*database.DB'
            $content = $content -replace 'models\.DB', 'database.DB'
            
            # Fix undefined: database references (missing package)
            # database.GetXXX -> models.GetXXX (for functions in models package)
            # But keep database.DB as is
            
            if ($content -ne $originalContent) {
                Set-Content -Path $file.FullName -Value $content -NoNewline
                Write-Host "    ✅ Updated" -ForegroundColor Green
            }
        }
    }
}

Write-Host "✅ All service files fixed!"

