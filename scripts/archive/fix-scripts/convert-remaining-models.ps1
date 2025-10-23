# Convert remaining model methods to functions

Write-Host "🔄 Converting remaining model methods..."

$file = "backend/content/models/tags.go"
$content = Get-Content $file -Raw

# Convert method signatures to function signatures
$content = $content -replace 'func \(db \*database\.DB\) ([A-Z][a-zA-Z0-9]+)\(', 'func $1(db *database.DB, '

# Fix cases where there are no additional parameters
$content = $content -replace 'func ([A-Z][a-zA-Z0-9]+)\(db \*database\.DB, \)', 'func $1(db *database.DB)'

# Fix internal function calls
$content = $content -replace 'db\.Get([A-Z][a-zA-Z0-9]+)\(', 'Get$1(db, '
$content = $content -replace 'db\.Create([A-Z][a-zA-Z0-9]+)\(', 'Create$1(db, '
$content = $content -replace 'db\.Update([A-Z][a-zA-Z0-9]+)\(', 'Update$1(db, '

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Converted tags.go!"

# Now fix video model tag function calls
$videoFile = "backend/video-streaming/models/master_video.go"
$videoContent = Get-Content $videoFile -Raw

# Import content models
if ($videoContent -notmatch 'contentModels') {
    $videoContent = $videoContent -replace '(import \()', '$1`n`tcontentModels "bome-backend/content/models"'
}

# Fix tag function calls to use contentModels
$videoContent = $videoContent -replace 'db\.convertTagWordsToIDs\(', 'convertTagWordsToIDs(db, '
$videoContent = $videoContent -replace 'UpdateTagFrequency\(', 'contentModels.UpdateTagFrequency(db, '
$videoContent = $videoContent -replace 'db\.convertTagIDsToWords\(', 'convertTagIDsToWords(db, '
$videoContent = $videoContent -replace 'db\.decrementTagFrequency\(', 'decrementTagFrequency(db, '
$videoContent = $videoContent -replace 'CreateTag\(', 'contentModels.CreateTag(db, '

Set-Content -Path $videoFile -Value $videoContent -NoNewline

Write-Host "✅ Fixed video tag calls!"

