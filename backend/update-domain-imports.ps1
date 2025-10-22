# Update all imports to use new domain-organized services

Write-Host "`n🔄 Updating imports across entire codebase...`n" -ForegroundColor Yellow

$files = Get-ChildItem -Path . -Include *.go -Recurse -File | Where-Object {
    $_.FullName -notlike "*\services\*"  # Don't update service files themselves
}

$count = 0
foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw
    $original = $content
    
    # Update security domain (crypto)
    $content = $content -replace 'bome-backend/services/crypto', 'bome-backend/services/security/crypto'
    $content = $content -replace '"bome-backend/ports"', '"bome-backend/services/security"'
    
    # Update payment domain (stripe)
    $content = $content -replace 'bome-backend/services/stripe', 'bome-backend/services/payment/stripe'
    
    # Update media domain (bunny)
    $content = $content -replace 'bome-backend/services/bunny', 'bome-backend/services/media/bunny'
    
    # Update communication domain (email)
    $content = $content -replace 'bome-backend/services/email', 'bome-backend/services/communication/email'
    
    # Update analytics domain
    $content = $content -replace 'bome-backend/services/analytics', 'bome-backend/services/analytics/subscription'
    
    # Only update if something changed
    if ($content -ne $original) {
        Set-Content -Path $file.FullName -Value $content -NoNewline
        $count++
        Write-Host "  ✅ Updated: $($file.Name)" -ForegroundColor Green
    }
}

Write-Host "`n✅ Updated $count files!`n" -ForegroundColor Green

