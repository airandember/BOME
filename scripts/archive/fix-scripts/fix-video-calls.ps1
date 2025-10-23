# Fix video streaming service calls

Write-Host "🔄 Fixing video streaming calls..."

$file = "backend/video-streaming/services/master_video_sync.go"
$content = Get-Content $file -Raw

# Fix GetMasterVideos call
$content = $content -replace 's\.db\.GetMasterVideos\(\)', 'videoModels.GetMasterVideos(s.db)'

# Replace database.MasterVideo with videoModels.MasterVideo
$content = $content -replace 'database\.MasterVideo', 'videoModels.MasterVideo'

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Video streaming fixed!"

