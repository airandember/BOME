# BOME PostgreSQL Database Setup Script
# This script sets up the PostgreSQL database for the BOME application

param(
    [string]$DatabaseName = "bome_db",
    [string]$Username = "bome_admin",
    [string]$Password = "AdminBOME",
    [string]$DbHost = "localhost",
    [int]$Port = 5432
)

Write-Host "=== BOME PostgreSQL Database Setup ===" -ForegroundColor Green
Write-Host ""

# Check if psql is available
try {
    $psqlVersion = psql --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ PostgreSQL client (psql) found: $psqlVersion" -ForegroundColor Green
    } else {
        throw "psql not found"
    }
} catch {
    Write-Host "✗ PostgreSQL client (psql) not found in PATH" -ForegroundColor Red
    Write-Host "Please install PostgreSQL and ensure psql is in your PATH" -ForegroundColor Yellow
    Write-Host "Download from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    exit 1
}

# Function to execute SQL command
function Invoke-PostgreSQL {
    param(
        [string]$Command,
        [string]$Database = "postgres"
    )
    
    $env:PGPASSWORD = $Password
    $result = psql -h $DbHost -p $Port -U $Username -d $Database -c $Command 2>&1
    $env:PGPASSWORD = ""
    
    return $result
}

# Function to execute SQL file
function Invoke-PostgreSQLFile {
    param(
        [string]$FilePath,
        [string]$Database = "postgres"
    )
    
    $env:PGPASSWORD = $Password
    $result = psql -h $DbHost -p $Port -U $Username -d $Database -f $FilePath 2>&1
    $env:PGPASSWORD = ""
    
    return $result
}

Write-Host "Step 1: Creating database user..." -ForegroundColor Yellow
try {
    # Try to connect as postgres superuser first
    $env:PGPASSWORD = "postgres"
    $testConnection = psql -h $DbHost -p $Port -U postgres -d postgres -c "SELECT 1;" 2>$null
    $env:PGPASSWORD = ""
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Connected as postgres superuser" -ForegroundColor Green
        
        # Create user
        $createUser = "CREATE USER $Username WITH PASSWORD '$Password';"
        $result = Invoke-PostgreSQL -Command $createUser -Database "postgres"
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ User '$Username' created successfully" -ForegroundColor Green
        } else {
            Write-Host "! User '$Username' may already exist (this is OK)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "! Could not connect as postgres superuser" -ForegroundColor Yellow
        Write-Host "  You may need to create the user manually or use existing credentials" -ForegroundColor Yellow
    }
} catch {
    Write-Host "! Error creating user: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Step 2: Creating database..." -ForegroundColor Yellow
try {
    # Create database
    $createDb = "CREATE DATABASE $DatabaseName OWNER $Username;"
    $result = Invoke-PostgreSQL -Command $createDb -Database "postgres"
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Database '$DatabaseName' created successfully" -ForegroundColor Green
    } else {
        Write-Host "! Database '$DatabaseName' may already exist (this is OK)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "! Error creating database: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Step 3: Setting up database schema..." -ForegroundColor Yellow

# Check if setup-database.sql exists
$schemaFile = Join-Path $PSScriptRoot "setup-database.sql"
if (Test-Path $schemaFile) {
    try {
        $result = Invoke-PostgreSQLFile -FilePath $schemaFile -Database $DatabaseName
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Database schema created successfully" -ForegroundColor Green
        } else {
            Write-Host "! Some schema operations may have failed (check output above)" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "! Error setting up schema: $_" -ForegroundColor Red
    }
} else {
    Write-Host "✗ Schema file not found: $schemaFile" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 4: Verifying setup..." -ForegroundColor Yellow

# Test connection to the new database
try {
    $testQuery = "SELECT COUNT(*) as table_count FROM information_schema.tables WHERE table_schema = 'public';"
    $result = Invoke-PostgreSQL -Command $testQuery -Database $DatabaseName
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Database connection test successful" -ForegroundColor Green
        Write-Host "  Tables found: $result" -ForegroundColor Green
    } else {
        Write-Host "✗ Database connection test failed" -ForegroundColor Red
    }
} catch {
    Write-Host "! Error testing connection: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Step 5: Creating .env file..." -ForegroundColor Yellow

# Create .env file with database configuration
$envFile = Join-Path $PSScriptRoot ".env"
$envContent = @"
# Database Configuration
DB_HOST=$DbHost
DB_PORT=$Port
DB_NAME=$DatabaseName
DB_USER=$Username
DB_PASSWORD=$Password

# JWT Configuration
JWT_SECRET=]d;SHv1;EL70)-l}ajibeNIKL>j$}:WD
JWT_REFRESH_SECRET=)KeV)cH8NStoq!4%6)xXt7MK7&)Xq*rX

# Server Configuration
PORT=8080
ENVIRONMENT=development

# Email Configuration (update with your email service)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Stripe Configuration (update with your Stripe keys)
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_PUBLISHABLE_KEY=pk_test_your_stripe_publishable_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

# Bunny.net Configuration (update with your Bunny.net credentials)
BUNNY_STORAGE_ZONE=your-storage-zone
BUNNY_API_KEY=your-api-key
BUNNY_REGION=de

# YouTube API Configuration (update with your YouTube API key)
YOUTUBE_API_KEY=your-youtube-api-key
"@

try {
    $envContent | Out-File -FilePath $envFile -Encoding UTF8
    Write-Host "✓ .env file created at: $envFile" -ForegroundColor Green
    Write-Host "  Please update the email, Stripe, Bunny.net, and YouTube API configurations" -ForegroundColor Yellow
} catch {
    Write-Host "! Error creating .env file: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== Setup Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Update the .env file with your actual API keys and credentials" -ForegroundColor White
Write-Host "2. Run the backend: go run main.go" -ForegroundColor White
Write-Host "3. Test the connection: go run cmd/create-admin/main.go" -ForegroundColor White
Write-Host ""
Write-Host "Database connection string:" -ForegroundColor Cyan
Write-Host "postgresql://$Username`:$Password@$DbHost`:$Port/$DatabaseName" -ForegroundColor White
Write-Host "" 