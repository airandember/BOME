# Fix user.go to convert methods to functions

Write-Host "🔄 Converting methods to functions in user.go..."

$file = "backend/authentication/models/user.go"
$content = Get-Content $file -Raw

# Convert method signatures to function signatures
# Pattern: func (db *database.DB) FunctionName( 
# Replace with: func FunctionName(db *database.DB,

$content = $content -replace 'func \(db \*database\.DB\) ([A-Z][a-zA-Z0-9]+)\(', 'func $1(db *database.DB, '

# Fix cases where there are no additional parameters (add closing paren)
$content = $content -replace 'func ([A-Z][a-zA-Z0-9]+)\(db \*database\.DB, \)', 'func $1(db *database.DB)'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Converted methods to functions in user.go"

