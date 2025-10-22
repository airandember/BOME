# Fix all malformed imports in backend

Write-Host "🔄 Fixing all malformed imports..."

$files = Get-ChildItem -Path "backend" -Recurse -Filter "*.go" | Where-Object { $_.FullName -notmatch "\\bin\\" }

$fixed = 0

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $originalContent = $content
    
    # Fix malformed imports (``n``t pattern)
    $content = $content -replace 'import \(``n``t"', 'import (`n`t"'
    
    if ($content -ne $originalContent) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
        $fixed++
        Write-Host "  ✅ Fixed: $($file.Name)"
    }
}

Write-Host "✅ Fixed $fixed files!"

