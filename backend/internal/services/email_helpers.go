package services

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

// sendTemplatedEmail sends an email using a template
func (s *EmailService) sendTemplatedEmail(to, subject, templateName string, data EmailData, emailType string, userID int) error {
	// Check if email is enabled
	enabled, err := s.isEmailEnabled()
	if err != nil {
		return fmt.Errorf("failed to check email settings: %w", err)
	}
	if !enabled {
		log.Printf("⚠️ [EMAIL] Email sending is disabled, skipping email to %s", to)
		return nil
	}

	// Get template
	tmpl, exists := s.templates[templateName]
	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}

	// Render template
	var htmlBody bytes.Buffer
	err = tmpl.Execute(&htmlBody, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Record email notification
	notificationID, err := s.recordEmailNotification(userID, to, emailType, subject, templateName, data)
	if err != nil {
		log.Printf("⚠️ [EMAIL] Failed to record notification: %v", err)
	}

	// Send with Resend (simplified - no more failover)
	sendErr := s.sendWithResend(to, subject, htmlBody.String(), notificationID)
	if sendErr != nil {
		log.Printf("❌ [EMAIL] Resend failed: %v", sendErr)
		if notificationID > 0 {
			s.updateEmailNotificationStatus(notificationID, "failed", sendErr.Error())
		}
		return fmt.Errorf("failed to send email via Resend: %w", sendErr)
	}

	log.Printf("✅ [EMAIL] Email sent successfully to %s", to)
	return nil
}

// sendWithResend sends email using Resend API with usage tracking
func (s *EmailService) sendWithResend(to, subject, htmlBody string, notificationID int) error {
	now := time.Now()
	today := now.Format("2006-01-02")
	year := now.Year()
	month := int(now.Month())

	// Get current usage
	dailyUsage, _ := s.getTodayUsage("resend", today)
	monthlyUsage, _ := s.getMonthlyUsage("resend", year, month)

	// Get limits (Resend free tier: 3000/month, 100/day)
	dailyLimit := s.getProviderLimit("resend", 100)
	monthlyLimit := s.getProviderMonthlyLimit("resend", 3000)

	// Check limits
	if dailyUsage >= dailyLimit {
		return fmt.Errorf("daily email limit reached for Resend (%d/%d)", dailyUsage, dailyLimit)
	}

	if monthlyUsage >= monthlyLimit {
		return fmt.Errorf("monthly email limit reached for Resend (%d/%d)", monthlyUsage, monthlyLimit)
	}

	log.Printf("📧 [RESEND] Sending email (daily: %d/%d, monthly: %d/%d)",
		dailyUsage, dailyLimit, monthlyUsage, monthlyLimit)

	// Send via Resend
	sendErr := s.sendResendEmail(to, subject, htmlBody)

	// Update usage tracking
	success := sendErr == nil
	s.incrementUsage("resend", success)

	if sendErr == nil {
		// Success! Update notification and return
		if notificationID > 0 {
			s.updateEmailNotificationStatus(notificationID, "sent", "Sent via Resend")
		}
		log.Printf("✅ [RESEND] Email sent successfully to %s", to)
		return nil
	}

	// Handle specific Resend errors
	if strings.Contains(sendErr.Error(), "domain is not verified") {
		return fmt.Errorf("resend domain verification required: %w", sendErr)
	}

	return fmt.Errorf("resend email failed: %w", sendErr)
}

// sendSMTPEmail sends an email via SMTP
func (s *EmailService) sendSMTPEmail(to, subject, htmlBody string) error {
	// Get SMTP settings
	settings, err := s.getSMTPSettings()
	if err != nil {
		return fmt.Errorf("failed to get SMTP settings: %w", err)
	}

	// Validate settings
	if settings.Host == "" || settings.Username == "" || settings.Password == "" || settings.FromEmail == "" {
		return fmt.Errorf("SMTP settings incomplete - please configure email settings")
	}

	// Create message
	message := fmt.Sprintf(
		"To: %s\r\n"+
			"From: %s <%s>\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		to, settings.FromName, settings.FromEmail, subject, htmlBody)

	// Setup authentication
	auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)

	// Setup TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         settings.Host,
	}

	// Connect and send
	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)

	// Create connection
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, settings.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender and recipient
	if err = client.Mail(settings.FromEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer writer.Close()

	_, err = writer.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Helper functions

func (s *EmailService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *EmailService) storeVerificationToken(userID int, token, email string, expiresAt time.Time) error {
	log.Printf("🔍 [TOKEN-STORE] Storing token for user %d: %s (expires: %v)", userID, token, expiresAt)

	query := `
		INSERT INTO email_verification_tokens (user_id, token, email, expires_at)
		VALUES ($1, $2, $3, $4)`

	_, err := s.db.DB.Exec(query, userID, token, email, expiresAt)
	if err != nil {
		log.Printf("❌ [TOKEN-STORE] Failed to store token: %v", err)
	} else {
		log.Printf("✅ [TOKEN-STORE] Token stored successfully")
	}
	return err
}

func (s *EmailService) getBaseURL() string {
	// Get from environment variable first
	if baseURL := os.Getenv("PUBLIC_APP_URL"); baseURL != "" {
		// Override for development - if it's pointing to frontend, use backend instead
		if strings.Contains(baseURL, "localhost:5173") {
			return "http://localhost:8080"
		}

		// Production fix: Add /bome-backend path to match frontend API calls
		// This ensures email verification URLs use the same routing as all other API calls
		if strings.Contains(baseURL, "bookofmormonevidence.org") && !strings.Contains(baseURL, "/bome-backend") {
			baseURL = baseURL + "/bome-backend"
		}

		return baseURL
	}

	// Try to get from database settings
	if baseURL, err := s.db.GetEmailSetting("app_base_url"); err == nil && baseURL != "" {
		return baseURL
	}

	// Development fallback with warning - use backend port for API endpoints
	return "http://localhost:8080"
}

// getFrontendBaseURL returns the frontend URL WITHOUT /bome-backend suffix
// Used for links that go directly to frontend pages (e.g., password reset, account settings)
func (s *EmailService) getFrontendBaseURL() string {
	// Get from environment variable first
	if baseURL := os.Getenv("PUBLIC_APP_URL"); baseURL != "" {
		// Development: frontend runs on port 5173
		if strings.Contains(baseURL, "localhost:8080") {
			return "http://localhost:5173"
		}
		if strings.Contains(baseURL, "localhost:5173") {
			return baseURL
		}

		// Production: strip /bome-backend if present, return just the domain
		baseURL = strings.TrimSuffix(baseURL, "/bome-backend")
		baseURL = strings.TrimSuffix(baseURL, "/")
		return baseURL
	}

	// Try to get from database settings
	if baseURL, err := s.db.GetEmailSetting("app_base_url"); err == nil && baseURL != "" {
		// Strip backend path if present
		baseURL = strings.TrimSuffix(baseURL, "/bome-backend")
		baseURL = strings.TrimSuffix(baseURL, "/")
		return baseURL
	}

	// Development fallback - frontend port
	return "http://localhost:5173"
}

func (s *EmailService) isEmailEnabled() (bool, error) {
	query := `SELECT setting_value FROM email_settings WHERE setting_key = 'email_enabled'`
	var value string
	err := s.db.DB.QueryRow(query).Scan(&value)
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

func (s *EmailService) getSMTPSettings() (*SMTPSettings, error) {
	settings := &SMTPSettings{}

	// Get all SMTP settings
	query := `
		SELECT setting_key, setting_value, is_encrypted 
		FROM email_settings 
		WHERE setting_key IN ('smtp_host', 'smtp_port', 'smtp_username', 'smtp_password', 'smtp_from_email', 'smtp_from_name')`

	rows, err := s.db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settingsMap := make(map[string]string)
	for rows.Next() {
		var key, value string
		var isEncrypted bool

		err := rows.Scan(&key, &value, &isEncrypted)
		if err != nil {
			continue
		}

		// Decrypt if needed
		if isEncrypted && value != "" && s.cryptoService != nil {
			decrypted, err := s.cryptoService.DecryptString(value)
			if err != nil {
				log.Printf("⚠️ [EMAIL] Failed to decrypt %s: %v", key, err)
				continue
			}
			value = decrypted
		}

		settingsMap[key] = value
	}

	// Map to struct
	settings.Host = settingsMap["smtp_host"]
	settings.Username = settingsMap["smtp_username"]
	settings.Password = settingsMap["smtp_password"]
	settings.FromEmail = settingsMap["smtp_from_email"]
	settings.FromName = settingsMap["smtp_from_name"]

	// Parse port
	if portStr := settingsMap["smtp_port"]; portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid SMTP port: %s", portStr)
		}
		settings.Port = port
	} else {
		settings.Port = 587 // Default
	}

	return settings, nil
}

func (s *EmailService) recordEmailNotification(userID int, emailTo, emailType, subject, templateName string, templateData EmailData) (int, error) {
	query := `
		INSERT INTO email_notifications (user_id, email_to, email_type, subject, template_name, template_data, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		RETURNING id`

	// Convert struct to JSON string for database storage
	templateDataJSON, err := json.Marshal(templateData)
	if err != nil {
		log.Printf("⚠️ [EMAIL] Failed to marshal template data: %v", err)
		templateDataJSON = []byte("{}")
	}

	var id int
	err = s.db.DB.QueryRow(query, userID, emailTo, emailType, subject, templateName, string(templateDataJSON)).Scan(&id)
	return id, err
}

func (s *EmailService) updateEmailNotificationStatus(id int, status, errorMessage string) error {
	query := `
		UPDATE email_notifications 
		SET status = $1, error_message = $2, sent_at = CASE WHEN $1 = 'sent' THEN CURRENT_TIMESTAMP ELSE sent_at END
		WHERE id = $3`

	_, err := s.db.DB.Exec(query, status, errorMessage, id)
	return err
}

// EmailProvider represents an email provider with usage tracking
type EmailProvider struct {
	Name        string
	DailyLimit  int
	UsedToday   int
	IsAvailable bool
}

// checkResendAvailability checks if Resend is available for sending
func (s *EmailService) checkResendAvailability() error {
	today := time.Now().Format("2006-01-02")
	year := time.Now().Year()
	month := int(time.Now().Month())

	// Get current usage
	dailyUsage, err := s.getTodayUsage("resend", today)
	if err != nil {
		log.Printf("⚠️ [EMAIL] Failed to get resend daily usage: %v", err)
		dailyUsage = 0 // Continue with 0 if we can't get usage
	}

	monthlyUsage, err := s.getMonthlyUsage("resend", year, month)
	if err != nil {
		log.Printf("⚠️ [EMAIL] Failed to get resend monthly usage: %v", err)
		monthlyUsage = 0 // Continue with 0 if we can't get usage
	}

	// Get limits from settings
	dailyLimit := s.getProviderLimit("resend", 100)
	monthlyLimit := s.getProviderMonthlyLimit("resend", 3000)

	// Check daily limit
	if dailyUsage >= dailyLimit {
		return fmt.Errorf("daily email limit reached for Resend (%d/%d)", dailyUsage, dailyLimit)
	}

	// Check monthly limit
	if monthlyUsage >= monthlyLimit {
		return fmt.Errorf("monthly email limit reached for Resend (%d/%d)", monthlyUsage, monthlyLimit)
	}

	log.Printf("📧 [EMAIL] Resend available (daily: %d/%d, monthly: %d/%d)",
		dailyUsage, dailyLimit, monthlyUsage, monthlyLimit)
	return nil
}

// getTodayUsage gets today's email usage for a specific provider
func (s *EmailService) getTodayUsage(provider, date string) (int, error) {
	var usage int
	query := `
		SELECT COALESCE(emails_sent, 0) 
		FROM daily_email_usage 
		WHERE date = $1 AND provider = $2`

	err := s.db.DB.QueryRow(query, date, provider).Scan(&usage)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get usage for %s: %w", provider, err)
	}
	return usage, nil
}

// getProviderLimit gets the daily limit for a provider from settings
func (s *EmailService) getProviderLimit(provider string, defaultLimit int) int {
	settingKey := fmt.Sprintf("daily_email_limit_%s", provider)
	limitStr, err := s.db.GetEmailSetting(settingKey)
	if err != nil || limitStr == "" {
		log.Printf("⚠️ [EMAIL-ROUTER] No limit set for %s, using default: %d", provider, defaultLimit)
		return defaultLimit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		log.Printf("⚠️ [EMAIL-ROUTER] Invalid limit for %s: %s, using default: %d", provider, limitStr, defaultLimit)
		return defaultLimit
	}

	return limit
}

// getProviderMonthlyLimit gets the monthly limit for a provider from settings
func (s *EmailService) getProviderMonthlyLimit(provider string, defaultLimit int) int {
	settingKey := fmt.Sprintf("monthly_email_limit_%s", provider)
	limitStr, err := s.db.GetEmailSetting(settingKey)
	if err != nil || limitStr == "" {
		log.Printf("⚠️ [EMAIL-ROUTER] No monthly limit set for %s, using default: %d", provider, defaultLimit)
		return defaultLimit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		log.Printf("⚠️ [EMAIL-ROUTER] Invalid monthly limit for %s: %s, using default: %d", provider, limitStr, defaultLimit)
		return defaultLimit
	}

	return limit
}

// getMonthlyUsage gets monthly email usage for a specific provider
func (s *EmailService) getMonthlyUsage(provider string, year, month int) (int, error) {
	var usage int
	query := `
		SELECT COALESCE(emails_sent, 0) 
		FROM monthly_email_usage 
		WHERE year = $1 AND month = $2 AND provider = $3`

	err := s.db.DB.QueryRow(query, year, month, provider).Scan(&usage)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return usage, nil
}

// incrementUsage updates usage statistics after sending an email
func (s *EmailService) incrementUsage(provider string, success bool) error {
	now := time.Now()
	today := now.Format("2006-01-02")
	year := now.Year()
	month := int(now.Month())

	// Update daily usage
	var dailyErr error
	if success {
		query := `
			INSERT INTO daily_email_usage (date, provider, emails_sent, emails_failed)
			VALUES ($1, $2, 1, 0)
			ON CONFLICT (date, provider) 
			DO UPDATE SET 
				emails_sent = daily_email_usage.emails_sent + 1,
				last_updated = CURRENT_TIMESTAMP`
		_, dailyErr = s.db.DB.Exec(query, today, provider)
		if dailyErr != nil {
			log.Printf("❌ [EMAIL-TRACKER] Failed to increment daily success count for %s: %v", provider, dailyErr)
		} else {
			log.Printf("📊 [EMAIL-TRACKER] Incremented daily success count for %s", provider)
		}
	} else {
		query := `
			INSERT INTO daily_email_usage (date, provider, emails_sent, emails_failed)
			VALUES ($1, $2, 0, 1)
			ON CONFLICT (date, provider) 
			DO UPDATE SET 
				emails_failed = daily_email_usage.emails_failed + 1,
				last_updated = CURRENT_TIMESTAMP`
		_, dailyErr = s.db.DB.Exec(query, today, provider)
		if dailyErr != nil {
			log.Printf("❌ [EMAIL-TRACKER] Failed to increment daily failure count for %s: %v", provider, dailyErr)
		} else {
			log.Printf("📊 [EMAIL-TRACKER] Incremented daily failure count for %s", provider)
		}
	}

	// Update monthly usage
	var monthlyErr error
	if success {
		query := `
			INSERT INTO monthly_email_usage (year, month, provider, emails_sent, emails_failed)
			VALUES ($1, $2, $3, 1, 0)
			ON CONFLICT (year, month, provider) 
			DO UPDATE SET 
				emails_sent = monthly_email_usage.emails_sent + 1,
				last_updated = CURRENT_TIMESTAMP`
		_, monthlyErr = s.db.DB.Exec(query, year, month, provider)
		if monthlyErr != nil {
			log.Printf("❌ [EMAIL-TRACKER] Failed to increment monthly success count for %s: %v", provider, monthlyErr)
		} else {
			log.Printf("📊 [EMAIL-TRACKER] Incremented monthly success count for %s", provider)
		}
	} else {
		query := `
			INSERT INTO monthly_email_usage (year, month, provider, emails_sent, emails_failed)
			VALUES ($1, $2, $3, 0, 1)
			ON CONFLICT (year, month, provider) 
			DO UPDATE SET 
				emails_failed = monthly_email_usage.emails_failed + 1,
				last_updated = CURRENT_TIMESTAMP`
		_, monthlyErr = s.db.DB.Exec(query, year, month, provider)
		if monthlyErr != nil {
			log.Printf("❌ [EMAIL-TRACKER] Failed to increment monthly failure count for %s: %v", provider, monthlyErr)
		} else {
			log.Printf("📊 [EMAIL-TRACKER] Incremented monthly failure count for %s", provider)
		}
	}

	// Return the first error encountered, or nil if both succeeded
	if dailyErr != nil {
		return dailyErr
	}
	return monthlyErr
}

// sendResendEmail sends an email via Resend API
func (s *EmailService) sendResendEmail(to, subject, htmlBody string) error {
	// Get Resend API key
	apiKey, err := s.db.GetEmailSetting("resend_api_key")
	if err != nil || apiKey == "" {
		return fmt.Errorf("resend API key not configured")
	}

	// Decrypt if encrypted
	if s.cryptoService != nil {
		decryptedKey, err := s.cryptoService.DecryptString(apiKey)
		if err == nil {
			apiKey = decryptedKey
		}
	}

	// Get sender email
	fromEmail, err := s.db.GetEmailSetting("smtp_from_email")
	if err != nil || fromEmail == "" {
		fromEmail = "noreply@bookofmormonevidence.org" // fallback
	}

	fromName, _ := s.db.GetEmailSetting("smtp_from_name")
	if fromName == "" {
		fromName = "BOME Support"
	}

	// Prepare Resend API request
	payload := map[string]interface{}{
		"from":    fmt.Sprintf("%s <%s>", fromName, fromEmail),
		"to":      []string{to},
		"subject": subject,
		"html":    htmlBody,
	}

	jsonPayload, _ := json.Marshal(payload)

	// Make HTTP request to Resend API
	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (%d): %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ [RESEND] Email sent successfully to %s", to)
	return nil
}
