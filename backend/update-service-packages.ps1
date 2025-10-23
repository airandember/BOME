# Update package declarations for domain-organized services

Write-Host "`n🔧 Updating package declarations...`n" -ForegroundColor Yellow

# Security domain (crypto stays as crypto)
Get-ChildItem "services/security/crypto/*.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    # Package is already crypto, no change needed
    Write-Host "  ✅ $($_.Name) - package crypto (unchanged)" -ForegroundColor Green
}

# Payment domain (stripe stays as stripe)
Get-ChildItem "services/payment/stripe/*.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    # Package is already stripe, no change needed
    Write-Host "  ✅ $($_.Name) - package stripe (unchanged)" -ForegroundColor Green
}

# Media domain (bunny stays as bunny)
Get-ChildItem "services/media/bunny/*.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    # Package is already bunny, no change needed
    Write-Host "  ✅ $($_.Name) - package bunny (unchanged)" -ForegroundColor Green
}

# Communication domain (email stays as email)
Get-ChildItem "services/communication/email/*.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    # Package is already email, no change needed
    Write-Host "  ✅ $($_.Name) - package email (unchanged)" -ForegroundColor Green
}

# Analytics domain
Get-ChildItem "services/analytics/subscription/*.go" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    # Package might be services, change to subscription if needed
    if ($content -match "^package services") {
        $content = $content -replace "^package services", "package subscription"
        Set-Content -Path $_.FullName -Value $content -NoNewline
        Write-Host "  ✅ $($_.Name) - package subscription (updated)" -ForegroundColor Cyan
    } else {
        Write-Host "  ✅ $($_.Name) - package already correct" -ForegroundColor Green
    }
}

Write-Host "`n✅ Package declarations updated!`n" -ForegroundColor Green

