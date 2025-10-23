# Add missing IsConfigured method to EmailService

$file = "services/email/email.go"
$content = Get-Content $file -Raw

# Find a good place to insert (after SendCustomEmail method or at the end)
# Insert before the final closing brace

$method = @"

// IsConfigured returns whether the email service is properly configured
func (s *EmailService) IsConfigured() bool {
	// Email service is configured if we have a database connection
	// In the future, this could check SMTP settings
	return s.db != nil
}
"@

# Insert before the last occurrence of a method or at the end of the file
$content = $content.TrimEnd() + "`n" + $method + "`n"

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Added IsConfigured method to EmailService" -ForegroundColor Green

