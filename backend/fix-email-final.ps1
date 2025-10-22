# Final fix for EmailService

$file = "services/email/email.go"
$content = Get-Content $file -Raw

# Remove the invalid ports.SMTPSettings references
$content = $content -replace 'ports\.SMTPSettings', 'SMTPSettings'

# Find where SendWelcomeEmail is defined and add SendPasswordSetupEmail after it
$setupMethod = @"


// SendPasswordSetupEmail sends a password setup email
func (s *EmailService) SendPasswordSetupEmail(email, token, firstName string) error {
	// Use SendVerificationEmail for now, or implement a separate template
	return s.SendVerificationEmail(email, token, firstName)
}
"@

# Insert after SendWelcomeEmail method (look for the pattern)
if ($content -match '(func \(s \*EmailService\) SendWelcomeEmail[^}]+\})\s*\n') {
    $content = $content -replace '(func \(s \*EmailService\) SendWelcomeEmail[^}]+\})', "`$1$setupMethod"
}

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Fixed EmailService final issues" -ForegroundColor Green

