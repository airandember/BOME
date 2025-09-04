package services

import (
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"bome-backend/internal/database"
)

// EmailService handles all email operations
type EmailService struct {
	db            *database.DB
	cryptoService *CryptoService
	templates     map[string]*template.Template
}

// EmailData represents data passed to email templates
type EmailData struct {
	UserName         string
	UserEmail        string
	VerificationURL  string
	SubscriptionName string
	Amount           string
	Currency         string
	PeriodEnd        string
	SupportEmail     string
	CompanyName      string
	BaseURL          string
}

// SMTPSettings holds SMTP configuration
type SMTPSettings struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

// NewEmailService creates a new email service
func NewEmailService(db *database.DB) *EmailService {
	service := &EmailService{
		db:            db,
		cryptoService: GetGlobalCryptoService(),
		templates:     make(map[string]*template.Template),
	}

	// Load email templates
	service.loadTemplates()

	return service
}

// loadTemplates loads all email templates
func (s *EmailService) loadTemplates() {
	// Email verification template
	verificationHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email - {{.CompanyName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2563eb; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 8px 8px; }
        .button { display: inline-block; background: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 14px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to {{.CompanyName}}!</h1>
    </div>
    <div class="content">
        <h2>Verify Your Email Address</h2>
        <p>Hi {{.UserName}},</p>
        <p>Thank you for signing up! Please verify your email address to complete your account setup.</p>
        <p><a href="{{.VerificationURL}}" class="button">Verify Email Address</a></p>
        <p>If the button doesn't work, copy and paste this link into your browser:</p>
        <p><a href="{{.VerificationURL}}">{{.VerificationURL}}</a></p>
        <p>This verification link will expire in 24 hours.</p>
    </div>
    <div class="footer">
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <p>Need help? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
    </div>
</body>
</html>`

	// Subscription confirmation template
	subscriptionHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Subscription Confirmed - {{.CompanyName}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #059669; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 8px 8px; }
        .success-icon { font-size: 48px; text-align: center; margin: 20px 0; }
        .amount { font-size: 24px; font-weight: bold; color: #059669; text-align: center; margin: 20px 0; }
        .details { background: white; padding: 20px; border-radius: 6px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 14px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🎉 Subscription Confirmed!</h1>
    </div>
    <div class="content">
        <div class="success-icon">✅</div>
        <h2>Welcome to {{.CompanyName}} Premium!</h2>
        <p>Hi {{.UserName}},</p>
        <p>Your subscription has been successfully activated. Thank you for supporting our mission!</p>
        
        <div class="details">
            <h3>Subscription Details:</h3>
            <p><strong>Plan:</strong> {{.SubscriptionName}}</p>
            <p><strong>Amount:</strong> <span class="amount">{{.Amount}} {{.Currency}}</span></p>
            <p><strong>Next billing:</strong> {{.PeriodEnd}}</p>
        </div>
        
        <p>You now have access to:</p>
        <ul>
            <li>Exclusive Book of Mormon evidence content</li>
            <li>Ad-free viewing experience</li>
            <li>Download videos for offline viewing</li>
            <li>Priority customer support</li>
            <li>Access to our community forum</li>
        </ul>
        
        <p><a href="{{.BaseURL}}/dashboard" class="button">Access Your Dashboard</a></p>
    </div>
    <div class="footer">
        <p>Questions? Contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
        <p>You can manage your subscription at any time from your account dashboard.</p>
    </div>
</body>
</html>`

	// Parse templates
	var err error
	s.templates["email_verification"], err = template.New("email_verification").Parse(verificationHTML)
	if err != nil {
		log.Printf("❌ [EMAIL] Failed to parse verification template: %v", err)
	}

	s.templates["subscription_confirmation"], err = template.New("subscription_confirmation").Parse(subscriptionHTML)
	if err != nil {
		log.Printf("❌ [EMAIL] Failed to parse subscription template: %v", err)
	}

	log.Printf("✅ [EMAIL] Loaded %d email templates", len(s.templates))
}

// SendVerificationEmail sends an email verification email
func (s *EmailService) SendVerificationEmail(userID int, email, name string) error {
	log.Printf("🔍 [EMAIL] Sending verification email to: %s", email)

	// Generate verification token
	token, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in database
	expiresAt := time.Now().Add(24 * time.Hour)
	err = s.storeVerificationToken(userID, token, email, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Get base URL from settings or use default
	baseURL := s.getBaseURL()
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)

	// Prepare email data
	emailData := EmailData{
		UserName:        name,
		UserEmail:       email,
		VerificationURL: verificationURL,
		CompanyName:     "BOME",
		SupportEmail:    "support@bome.test",
		BaseURL:         baseURL,
	}

	// Send email
	return s.sendTemplatedEmail(
		email,
		"Verify Your Email - BOME",
		"email_verification",
		emailData,
		"verification",
		userID,
	)
}

// SendSubscriptionConfirmation sends a subscription confirmation email
func (s *EmailService) SendSubscriptionConfirmation(userID int, email, name, subscriptionName string, amount float64, currency, periodEnd string) error {
	log.Printf("🔍 [EMAIL] Sending subscription confirmation to: %s", email)

	// Format amount
	amountStr := fmt.Sprintf("%.2f", amount)

	// Prepare email data
	emailData := EmailData{
		UserName:         name,
		UserEmail:        email,
		SubscriptionName: subscriptionName,
		Amount:           amountStr,
		Currency:         strings.ToUpper(currency),
		PeriodEnd:        periodEnd,
		CompanyName:      "BOME",
		SupportEmail:     "support@bome.test",
		BaseURL:          s.getBaseURL(),
	}

	// Send email
	return s.sendTemplatedEmail(
		email,
		"Subscription Confirmed - BOME",
		"subscription_confirmation",
		emailData,
		"subscription_confirmation",
		userID,
	)
}

// VerifyEmail verifies an email using a token
func (s *EmailService) VerifyEmail(token string) error {
	log.Printf("🔍 [EMAIL] Verifying email token: %s", token[:8]+"...")

	// Get token from database
	query := `
		SELECT user_id, email, expires_at, verified_at 
		FROM email_verification_tokens 
		WHERE token = $1`

	var userID int
	var email string
	var expiresAt time.Time
	var verifiedAt *time.Time

	err := s.db.DB.QueryRow(query, token).Scan(&userID, &email, &expiresAt, &verifiedAt)
	if err != nil {
		return fmt.Errorf("invalid or expired token")
	}

	// Check if already verified
	if verifiedAt != nil {
		return fmt.Errorf("email already verified")
	}

	// Check if expired
	if time.Now().After(expiresAt) {
		return fmt.Errorf("verification token has expired")
	}

	// Mark token as verified
	updateTokenQuery := `
		UPDATE email_verification_tokens 
		SET verified_at = CURRENT_TIMESTAMP 
		WHERE token = $1`

	_, err = s.db.DB.Exec(updateTokenQuery, token)
	if err != nil {
		return fmt.Errorf("failed to mark token as verified: %w", err)
	}

	// Update user as email verified
	updateUserQuery := `
		UPDATE users 
		SET email_verified = true, email_verified_at = CURRENT_TIMESTAMP 
		WHERE id = $1`

	_, err = s.db.DB.Exec(updateUserQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to mark user as verified: %w", err)
	}

	log.Printf("✅ [EMAIL] Email verified successfully for user %d", userID)
	return nil
}

// SendPasswordResetEmail sends a password reset email (placeholder for now)
func (s *EmailService) SendPasswordResetEmail(name, email, token string) error {
	log.Printf("🔍 [EMAIL] Password reset email requested for: %s", email)

	// TODO: Implement password reset email template and sending
	// For now, just log the request
	log.Printf("📧 [EMAIL] Would send password reset email to %s with token %s", email, token[:8]+"...")

	return nil
}
