# Comb All User Management Strands

Write-Host ""
Write-Host "USER MANAGEMENT BRAID - COMPLETE COMBING" -ForegroundColor Blue
Write-Host "=========================================" -ForegroundColor Blue
Write-Host ""

# Check if user-management directory exists
if (-not (Test-Path "user-management")) {
    Write-Host "[CRITICAL] user-management directory does not exist!" -ForegroundColor Red
    Write-Host "This might be in authentication or admin packages..." -ForegroundColor Yellow
    Write-Host ""
}

# STRAND 1: User Profile Management
Write-Host "STRAND 1: User Profile Management" -ForegroundColor Yellow
Write-Host "----------------------------------" -ForegroundColor Yellow

$hasProfileHandler = Select-String -Path "**/*.go" -Pattern "func.*Profile|func.*UpdateProfile|func.*GetProfile" -Quiet 2>$null
if ($hasProfileHandler) { 
    Write-Host "  [OK] Profile handlers found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Profile handlers missing" -ForegroundColor Red 
}

Write-Host ""

# STRAND 2: RBAC
Write-Host "STRAND 2: Role-Based Access Control (RBAC)" -ForegroundColor Yellow
Write-Host "-------------------------------------------" -ForegroundColor Yellow

$hasRBAC = Select-String -Path "**/*.go" -Pattern "func.*Role|func.*Permission|func.*CheckAccess" -Quiet 2>$null
if ($hasRBAC) { 
    Write-Host "  [OK] RBAC functions found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] RBAC missing" -ForegroundColor Red 
}

# Check for role middleware
$hasRoleMiddleware = Select-String -Path "**/*.go" -Pattern "RoleRequired|RequireRole|AdminOnly" -Quiet 2>$null
if ($hasRoleMiddleware) { 
    Write-Host "  [OK] Role middleware found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Role middleware missing" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 3: User Preferences
Write-Host "STRAND 3: User Preferences and Settings" -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor Yellow

$hasPreferences = Select-String -Path "**/*.go" -Pattern "func.*Preferences|func.*Settings" -Quiet 2>$null
if ($hasPreferences) { 
    Write-Host "  [OK] Preferences handlers found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Preferences missing" -ForegroundColor Yellow 
}

Write-Host ""

# STRAND 4: Admin User Management
Write-Host "STRAND 4: Admin User Management" -ForegroundColor Yellow
Write-Host "-------------------------------" -ForegroundColor Yellow

# Check admin package
if (Test-Path "admin/handlers/*.go") {
    Write-Host "  [OK] Admin handlers directory exists" -ForegroundColor Green
    
    $hasUserManagement = Select-String -Path "admin/handlers/*.go" -Pattern "func.*User|func.*GetUser|func.*UpdateUser|func.*DeleteUser" -Quiet
    if ($hasUserManagement) {
        Write-Host "  [OK] Admin user management found" -ForegroundColor Green
    } else {
        Write-Host "  [SPLIT-END] Admin user management missing" -ForegroundColor Red
    }
} else {
    Write-Host "  [SPLIT-END] Admin handlers directory missing" -ForegroundColor Red
}

Write-Host ""

# STRAND 5: Activity Tracking
Write-Host "STRAND 5: User Activity Tracking" -ForegroundColor Yellow
Write-Host "--------------------------------" -ForegroundColor Yellow

$hasActivityTracking = Select-String -Path "**/*.go" -Pattern "func.*Activity|func.*AuditLog|func.*TrackUser" -Quiet 2>$null
if ($hasActivityTracking) { 
    Write-Host "  [OK] Activity tracking found" -ForegroundColor Green 
} else { 
    Write-Host "  [SPLIT-END] Activity tracking missing" -ForegroundColor Yellow 
}

Write-Host ""

# Check for compilation errors
Write-Host "CHECKING FOR KNOWN ISSUES:" -ForegroundColor Cyan
Write-Host "  Checking authentication models (user profile data)..." -ForegroundColor Gray

if (Test-Path "authentication/models/user.go") {
    Write-Host "  [OK] User model exists in authentication" -ForegroundColor Green
} else {
    Write-Host "  [SPLIT-END] User model missing" -ForegroundColor Red
}

Write-Host ""
Write-Host "ALL 5 STRANDS COMBED" -ForegroundColor Green
Write-Host ""

