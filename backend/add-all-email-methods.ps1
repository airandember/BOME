# Add all missing methods to EmailService

$file = "services/email/email.go"
$content = Get-Content $file -Raw

# Add all missing methods in one go

$methods = @"

// IsValidEmail validates an email address format
func (s *EmailService) IsValidEmail(email string) bool {
	// Use the crypto service for validation
	if s.cryptoService == nil {
		return false
	}
	return s.cryptoService.ValidateEmail(email) == nil
}

// GetSMTPSettings retrieves the current SMTP configuration
func (s *EmailService) GetSMTPSettings() (*ports.SMTPSettings, error) {
	// Return default settings or load from database
	// For now, return nil to indicate not configured
	return nil, fmt.Errorf("SMTP settings not configured")
}

// UpdateSMTPSettings updates the SMTP configuration
func (s *EmailService) UpdateSMTPSettings(settings *ports.SMTPSettings) error {
	// Store SMTP settings in database
	// For now, return not implemented
	return fmt.Errorf("SMTP settings update not implemented")
}
"@

# Insert before the last occurrence of a method or at the end of the file
$content = $content.TrimEnd() + "`n" + $methods + "`n"

Set-Content -Path $file -Value $content -NoNewline

Write-Host "✅ Added all missing methods to EmailService" -ForegroundColor Green

