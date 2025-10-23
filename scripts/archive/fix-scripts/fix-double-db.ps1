# Fix double db parameters

Write-Host "🔄 Fixing double db parameters..."

$file = "backend/video-streaming/models/master_video.go"
$content = Get-Content $file -Raw

# Fix double db parameters
$content = $content -replace 'updateTagFrequency\(db, db,', 'updateTagFrequency(db,'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Fixed!"

