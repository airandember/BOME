# Fix middleware imports properly

$file = "authentication/middleware/middleware.go"
$lines = Get-Content $file

# Find the import block and fix it
$newLines = @()
$inImport = $false

foreach ($line in $lines) {
    if ($line -match "^import \(") {
        $inImport = $true
        $newLines += $line
        continue
    }
    
    if ($inImport -and $line -match "^\)") {
        $inImport = $false
    }
    
    # Skip malformed content models line
    if ($line -match "^\s*contentModels\s*$") {
        continue
    }
    
    $newLines += $line
}

$newLines | Set-Content -Path $file
Write-Host "✅ Fixed middleware imports" -ForegroundColor Green

