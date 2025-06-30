# Generate secure JWT secrets for BOME backend
Add-Type -AssemblyName System.Web

# Generate JWT secrets
$jwtSecret = [System.Web.Security.Membership]::GeneratePassword(32, 8)
$jwtRefreshSecret = [System.Web.Security.Membership]::GeneratePassword(32, 8)

Write-Host "Generated JWT Secrets:"
Write-Host "JWT_SECRET: $jwtSecret"
Write-Host "JWT_REFRESH_SECRET: $jwtRefreshSecret"
Write-Host ""

# Update .env file
$envContent = Get-Content .env -Raw

# Replace JWT secrets
$envContent = $envContent -replace 'JWT_SECRET=.*', "JWT_SECRET=$jwtSecret"
$envContent = $envContent -replace 'JWT_REFRESH_SECRET=.*', "JWT_REFRESH_SECRET=$jwtRefreshSecret"

# Update admin password
$adminPassword = [System.Web.Security.Membership]::GeneratePassword(16, 6)
$envContent = $envContent -replace 'ADMIN_PASSWORD=.*', "ADMIN_PASSWORD=$adminPassword"

# Update other security keys
$sessionSecret = [System.Web.Security.Membership]::GeneratePassword(32, 8)
$csrfSecret = [System.Web.Security.Membership]::GeneratePassword(32, 8)
$adminSecretKey = [System.Web.Security.Membership]::GeneratePassword(32, 8)

$envContent = $envContent -replace 'SESSION_SECRET=.*', "SESSION_SECRET=$sessionSecret"
$envContent = $envContent -replace 'CSRF_SECRET=.*', "CSRF_SECRET=$csrfSecret"
$envContent = $envContent -replace 'ADMIN_SECRET_KEY=.*', "ADMIN_SECRET_KEY=$adminSecretKey"

# Save updated .env file
Set-Content .env $envContent

Write-Host "Updated .env file with secure secrets!"
Write-Host "Admin Password: $adminPassword"
Write-Host ""
Write-Host "You can now start the server with: .\bome-backend.exe" 