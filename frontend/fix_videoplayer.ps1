# Create a PowerShell script to fix the VideoPlayer component
$content = Get-Content "src/lib/components/VideoPlayer.svelte" -Raw

# Remove the extractDirectVideoUrl function and its usage
$content = $content -replace '(?s)    // Also try to extract direct video URL from iframe if available.*?directVideoUrl = iframeSrc \? extractDirectVideoUrl\(iframeSrc\) : '''';', ''
$content = $content -replace '(?s)    function extractDirectVideoUrl\(iframeSrc: string\): string \{.*?\}', ''

# Update the logic to use iframeSrc directly as the direct video URL
$content = $content -replace 'directVideoUrl', 'iframeSrc'

# Remove unused references to directVideoUrl in error handling
$content = $content -replace 'if \(directVideoUrl\) \{\s*useDirectVideo\(\);\s*\} else \{\s*switchToIframe\(\);\s*\}', 'switchToIframe();'

# Update the useDirectVideo function to use iframeSrc
$content = $content -replace 'function useDirectVideo\(\) \{\s*if \(videoElement && directVideoUrl\) \{\s*console\.log\(''Setting direct video source:'', directVideoUrl\);\s*videoElement\.src = directVideoUrl;', 'function useDirectVideo() {
        if (videoElement && iframeSrc) {
            console.log(''Setting direct video source:'', iframeSrc);
            videoElement.src = iframeSrc;'

# Save the fixed content
$content | Set-Content "src/lib/components/VideoPlayer_fixed.svelte"
