# Fix remaining video model issues

Write-Host "🔄 Fixing video model..."

$file = "backend/video-streaming/models/master_video.go"
$content = Get-Content $file -Raw

# Fix UpdateTagFrequency calls
$content = $content -replace 'contentModels\.UpdateTagFrequency\(db,', 'updateTagFrequency(db,'

# Fix CreateTag call (remove extra db parameter)
$content = $content -replace 'contentModels\.CreateTag\(db, db,', 'contentModels.CreateTag(db,'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Video model fixed!"

