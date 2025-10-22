# Convert all model methods to functions

Write-Host "🔄 Converting model methods to functions..."

$modelDirs = @(
    "backend/subscription/models",
    "backend/video-streaming/models"
)

foreach ($dir in $modelDirs) {
    if (Test-Path $dir) {
        $files = Get-ChildItem -Path $dir -Filter "*.go"
        foreach ($file in $files) {
            Write-Host "  Processing: $($file.Name)" -ForegroundColor Gray
            
            $content = Get-Content $file.FullName -Raw
            $originalContent = $content
            
            # Convert method signatures to function signatures
            $content = $content -replace 'func \(db \*database\.DB\) ([A-Z][a-zA-Z0-9]+)\(', 'func $1(db *database.DB, '
            
            # Fix cases where there are no additional parameters
            $content = $content -replace 'func ([A-Z][a-zA-Z0-9]+)\(db \*database\.DB, \)', 'func $1(db *database.DB)'
            
            # Fix internal function calls (db.GetXXX -> GetXXX(db,)
            $content = $content -replace 'db\.Get([A-Z][a-zA-Z0-9]+)\(', 'Get$1(db, '
            $content = $content -replace 'db\.Create([A-Z][a-zA-Z0-9]+)\(', 'Create$1(db, '
            $content = $content -replace 'db\.Update([A-Z][a-zA-Z0-9]+)\(', 'Update$1(db, '
            $content = $content -replace 'db\.Delete([A-Z][a-zA-Z0-9]+)\(', 'Delete$1(db, '
            $content = $content -replace 'db\.Set([A-Z][a-zA-Z0-9]+)\(', 'Set$1(db, '
            
            if ($content -ne $originalContent) {
                Set-Content -Path $file.FullName -Value $content -NoNewline
                Write-Host "    ✅ Updated" -ForegroundColor Green
            }
        }
    }
}

Write-Host "✅ All model methods converted!"

