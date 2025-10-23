# Fix all model files to import and use database.DB

Write-Host "🔄 Fixing all model files..."

$modelDirs = @(
    "backend/authentication/models",
    "backend/video-streaming/models",
    "backend/subscription/models",
    "backend/user-management/models",
    "backend/analytics/models",
    "backend/advertisement/models",
    "backend/content/models"
)

foreach ($dir in $modelDirs) {
    if (Test-Path $dir) {
        $files = Get-ChildItem -Path $dir -Filter "*.go"
        foreach ($file in $files) {
            Write-Host "  Processing: $($file.Name)" -ForegroundColor Gray
            
            $content = Get-Content $file.FullName -Raw
            $originalContent = $content
            
            # Add import if not present
            if ($content -notmatch '"bome-backend/infrastructure/database"') {
                $content = $content -replace '(import \()', '$1`n`t"bome-backend/infrastructure/database"'
            }
            
            # Replace *DB, DB) references
            $content = $content -replace '\*DB,', '*database.DB,'
            $content = $content -replace '\(db \*DB\)', '(db *database.DB)'
            $content = $content -replace 'func \([^)]+\) ([^(]+)\(db \*DB', 'func ($1) $2(db *database.DB'
            
            if ($content -ne $originalContent) {
                Set-Content -Path $file.FullName -Value $content -NoNewline
                Write-Host "    ✅ Updated" -ForegroundColor Green
            }
        }
    }
}

Write-Host "✅ All model files fixed!"

