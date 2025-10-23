package email

import (
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	authModels "bome-backend/authentication/models"
	"bome-backend/infrastructure/database"
	cryptoSvc "bome-backend/services/security/crypto"
)

// Ensure EmailService implements ports.EmailPort
// TODO: Fix method signatures to fully conform to ports.EmailPort
// var _ ports.EmailPort = (*EmailService)(nil)

// EmailService handles all email operations
type EmailService struct {
	db            *database.DB
	cryptoService *cryptoSvc.CryptoService
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
		cryptoService: cryptoSvc.GetGlobalCryptoService(),
		templates:     make(map[string]*template.Template),
	}

	// Load email templates
	service.loadTemplates()

	return service
}

// isDevelopmentMode checks if we're in development mode
func (s *EmailService) isDevelopmentMode() bool {
	// Check if we're running on localhost or development environment
	baseURL := s.getBaseURL()
	return strings.Contains(baseURL, "localhost") || strings.Contains(baseURL, "127.0.0.1") || strings.Contains(baseURL, "bome.test")
}

// sendMockVerificationEmail simulates sending a verification email in development
func (s *EmailService) sendMockVerificationEmail(userID int, email, name string) error {
	// Generate verification token for database consistency
	token, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in database (use the same method as production)
	if err := authModels.SetVerificationToken(s.db, userID, token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Get base URL from settings or use default
	baseURL := s.getBaseURL()
	verificationURL := fmt.Sprintf("%s/api/v1/auth/verify-email-link?token=%s&user_id=%d", baseURL, token, userID)

	// Log the mock email instead of sending
	log.Printf("📧 [MOCK-EMAIL] ==================== VERIFICATION EMAIL ====================")
	log.Printf("📧 [MOCK-EMAIL] To: %s", email)
	log.Printf("📧 [MOCK-EMAIL] Subject: 🔐 Verify Your Email - Book of Mormon Evidence")
	log.Printf("📧 [MOCK-EMAIL] ")
	log.Printf("📧 [MOCK-EMAIL] 📖 Welcome to BOME!")
	log.Printf("📧 [MOCK-EMAIL] Book of Mormon Evidence Platform")
	log.Printf("📧 [MOCK-EMAIL] ")
	log.Printf("📧 [MOCK-EMAIL] Hi %s,", name)
	log.Printf("📧 [MOCK-EMAIL] ")
	log.Printf("📧 [MOCK-EMAIL] Welcome to the Book of Mormon Evidence community! We're excited")
	log.Printf("📧 [MOCK-EMAIL] to have you join thousands of others exploring the historical")
	log.Printf("📧 [MOCK-EMAIL] and archaeological evidence supporting the Book of Mormon.")
	log.Printf("📧 [MOCK-EMAIL] ")
	log.Printf("📧 [MOCK-EMAIL] 🔗 VERIFICATION LINK:")
	log.Printf("📧 [MOCK-EMAIL] %s", verificationURL)
	log.Printf("📧 [MOCK-EMAIL] ")
	log.Printf("📧 [MOCK-EMAIL] ⏰ This verification link will expire in 3 hours.")
	log.Printf("📧 [MOCK-EMAIL] ")
	log.Printf("📧 [MOCK-EMAIL] Once verified, you'll have access to:")
	log.Printf("📧 [MOCK-EMAIL] 📚 Exclusive research articles and studies")
	log.Printf("📧 [MOCK-EMAIL] 🎥 Premium video content and documentaries")
	log.Printf("📧 [MOCK-EMAIL] 🗺️ Interactive maps and archaeological findings")
	log.Printf("📧 [MOCK-EMAIL] 👥 Community discussions and expert insights")
	log.Printf("📧 [MOCK-EMAIL] 📱 Mobile app access for offline reading")
	log.Printf("📧 [MOCK-EMAIL] ")
	log.Printf("📧 [MOCK-EMAIL] Need help? Contact: support@bookofmormonevidence.org")
	log.Printf("📧 [MOCK-EMAIL] ================================================================")

	log.Printf("✅ [MOCK-EMAIL] Verification email logged for %s (token: %s)", email, token[:8]+"...")
	return nil
}

// loadTemplates loads all email templates
func (s *EmailService) loadTemplates() {
	// Email verification template - Professional BOME design
	verificationHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email - Book of Mormon Evidence</title>
    <!--[if mso]>
    <noscript>
        <xml>
            <o:OfficeDocumentSettings>
                <o:PixelsPerInch>96</o:PixelsPerInch>
            </o:OfficeDocumentSettings>
        </xml>
    </noscript>
    <![endif]-->
    <style>
        /* Reset and base styles */
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif; 
            line-height: 1.6; 
            color: #1f2937; 
            background-color: #f9fafb;
            margin: 0;
            padding: 0;
        }
        
        /* Container */
        .email-container { 
            max-width: 600px; 
            margin: 0 auto; 
            background-color: #ffffff;
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
        }
        
        /* Header with gradient */
        .header { 
            background: linear-gradient(135deg, #a27c4b 0%,#223858  100%);
            color: white; 
            padding: 40px 30px; 
            text-align: center;
            position: relative;
            overflow: hidden;
        }
        .header::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><pattern id="grain" width="100" height="100" patternUnits="userSpaceOnUse"><circle cx="25" cy="25" r="1" fill="white" opacity="0.1"/><circle cx="75" cy="75" r="1" fill="white" opacity="0.1"/><circle cx="50" cy="10" r="0.5" fill="white" opacity="0.1"/><circle cx="10" cy="60" r="0.5" fill="white" opacity="0.1"/><circle cx="90" cy="40" r="0.5" fill="white" opacity="0.1"/></pattern></defs><rect width="100" height="100" fill="url(%23grain)"/></svg>');
            pointer-events: none;
        }
        .header h1 { 
            font-size: 28px; 
            font-weight: 700; 
            margin-bottom: 8px;
            position: relative;
            z-index: 1;
        }
        .header p { 
            font-size: 16px; 
            opacity: 0.9;
            position: relative;
            z-index: 1;
        }
        
        /* Logo/Icon */
        .logo { 
            width: 64px; 
            height: 64px; 
            background: rgba(255, 255, 255, 0.2);
            border-radius: 50%;
            margin: 0 auto 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 32px;
            position: relative;
            z-index: 1;
        }
        
        /* Content area */
        .content { 
            padding: 40px 30px;
            background: #ffffff;
        }
        .content h2 { 
            color: #1f2937; 
            font-size: 24px; 
            font-weight: 600; 
            margin-bottom: 20px;
            text-align: center;
        }
        .content p { 
            color: #4b5563; 
            font-size: 16px; 
            margin-bottom: 16px;
            line-height: 1.7;
        }
        .greeting { 
            font-size: 18px; 
            color: #1f2937; 
            font-weight: 500;
        }
        
        /* Call-to-action button */
        .cta-container { 
            text-align: center; 
            margin: 32px 0;
        }
        .button { 
            display: inline-block; 
            background: linear-gradient(135deg,#a27c4b  0%,#223858 100%);
            color: white; 
            padding: 16px 32px; 
            text-decoration: none; 
            border-radius: 8px; 
            font-weight: 600;
            font-size: 16px;
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
            transition: all 0.3s ease;
        }
        .button:hover { 
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
        }
        
        /* Fallback link */
        .fallback-link { 
            background: #f3f4f6; 
            border-radius: 8px; 
            padding: 20px; 
            margin: 24px 0;
            border-left: 4px solid #223858;
        }
        .fallback-link p { 
            margin-bottom: 8px; 
            font-size: 14px;
            color: #6b7280;
        }
        .fallback-link a { 
            color: #223858; 
            word-break: break-all;
            font-size: 14px;
        }
        
        /* Security notice */
        .security-notice { 
            background: #fef3c7; 
            border-radius: 8px; 
            padding: 16px; 
            margin: 24px 0;
            border-left: 4px solid #f59e0b;
        }
        .security-notice p { 
            color: #92400e; 
            font-size: 14px; 
            margin-bottom: 0;
        }
        
        /* Footer */
        .footer { 
            background: #f9fafb; 
            padding: 30px; 
            text-align: center;
            border-top: 1px solid #e5e7eb;
        }
        .footer p { 
            color: #6b7280; 
            font-size: 14px; 
            margin-bottom: 8px;
        }
        .footer a { 
            color:#223858; 
            text-decoration: none;
        }
        .footer a:hover { 
            text-decoration: underline;
        }
        
        /* Company info */
        .company-info { 
            margin-top: 20px; 
            padding-top: 20px; 
            border-top: 1px solid #e5e7eb;
        }
        .company-info p { 
            font-size: 12px; 
            color: #9ca3af;
        }
        
        /* Mobile responsiveness */
        @media only screen and (max-width: 600px) {
            .email-container { margin: 0 10px; }
            .header { padding: 30px 20px; }
            .header h1 { font-size: 24px; }
            .content { padding: 30px 20px; }
            .content h2 { font-size: 20px; }
            .button { padding: 14px 24px; font-size: 15px; }
            .footer { padding: 20px; }
        }
        
        /* Dark mode support */
        @media (prefers-color-scheme: dark) {
            .content { background: #1f2937; }
            .content h2 { color: #f9fafb; }
            .content p { color: #d1d5db; }
            .content a { color: #d1d5db }
            .security-notice p { color: black }
            ul li { color: rgb(255, 242, 205) }
            .button {box-shadow: 0 0 0 black}
            .button:hover { box-shadow: 0 6px 20px rgba(133, 97, 1, 0.6) }
            .greeting { color: #f9fafb; }
            .fallback-link { background: #374151; }
            .fallback-link p { color: #9ca3af; }
        } 
    </style>
</head>
<body>
    <div class="email-container">
        <!-- Header -->
        <div class="header">
            <div class="logo">📖</div>
            <h1>Welcome to<br> Book of Mormon Evidence!</h1>
            <p>Book of Mormon Evidence Beta</p>
        </div>
        
        <!-- Main Content -->
        <div class="content">
            <h2>🔐 Verify Your Email Address</h2>
            <p class="greeting">Hi!</p>
            <p>Welcome to the Book of Mormon Evidence community! We're excited to have you join thousands of others exploring the historical and archaeological evidence supporting the Book of Mormon, and other vast topics by amazing speakers.</p>
            <p>To complete your account setup and access our exclusive content, please verify your email address by clicking the button below:</p>
            
            <div class="cta-container">
                <a href="{{.VerificationURL}}" class="button">✅ Verify Email Address</a>
            </div>
            
            <div class="fallback-link">
                <p><strong>Button not working?</strong> Copy and paste this link into your browser:</p>
                <a href="{{.VerificationURL}}">{{.VerificationURL}}</a>
            </div>
            
            <div class="security-notice">
                <p>⏰ <strong>Security Notice:</strong> This verification link will expire in 3 hours for your account security.</p>
            </div>
            
            <p>Once verified, you'll have access to:</p>
            <ul style="color: #4b5563; margin-left: 20px; margin-bottom: 20px;">
                <li>📚 Exclusive research articles and studies</li>
                <li>🎥 Premium video content and documentaries</li>
                <li>🗺️ Interactive maps and archaeological findings</li>
                <li>👥 Community discussions and expert insights</li>
                <li>📱 Mobile app access for offline reading</li>
            </ul>
        </div>
        
        <!-- Footer -->
        <div class="footer">
            <p>If you didn't create an account with us, you can safely ignore this email.</p>
            <p>Need help? Contact our support team at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a></p>
            
            <div class="company-info">
                <p><strong>Book of Mormon Evidence (BOME)</strong></p>
                <p>Strengthening faith through evidence-based research</p>
                <p>© 2024 Book of Mormon Evidence. All rights reserved.</p>
            </div>
        </div>
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

	// In development mode, just log the email instead of sending
	if s.isDevelopmentMode() {
		return s.sendMockVerificationEmail(userID, email, name)
	}

	// Generate verification token
	token, err := s.generateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in database (use the same method as mock)
	if err := authModels.SetVerificationToken(s.db, userID, token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Get base URL from settings or use default
	baseURL := s.getBaseURL()
	// Use backend API endpoint for verification (will redirect to frontend)
	verificationURL := fmt.Sprintf("%s/api/v1/auth/verify-email-link?token=%s&user_id=%d", baseURL, token, userID)

	// Prepare email data
	// Get support email from database
	supportEmail, err := s.db.GetEmailSetting("support_email")
	if err != nil || supportEmail == "" {
		supportEmail = "support@bookofmormonevidence.org" // fallback
	}

	emailData := EmailData{
		UserName:        name,
		UserEmail:       email,
		VerificationURL: verificationURL,
		CompanyName:     "BOME",
		SupportEmail:    supportEmail,
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

// SendTestEmail sends a test email to verify email configuration
func (s *EmailService) SendTestEmail(to, subject, body string, userID int) error {
	log.Printf("🔍 [EMAIL] Sending test email to: %s", to)

	// Check if email is enabled
	enabled, err := s.isEmailEnabled()
	if err != nil {
		return fmt.Errorf("failed to check email settings: %w", err)
	}
	if !enabled {
		return fmt.Errorf("email sending is disabled")
	}

	// Create simple HTML body if plain text provided
	htmlBody := body
	if !strings.Contains(body, "<html>") && !strings.Contains(body, "<p>") {
		htmlBody = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #3b82f6; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 8px 8px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 14px; color: #666; }
    </style>
</head>
<body>
    <div class="header">
        <h1>📧 Test Email</h1>
    </div>
    <div class="content">
        <p>%s</p>
        <p><strong>Sent at:</strong> %s</p>
    </div>
    <div class="footer">
        <p>This is a test email from BOME Admin Dashboard.</p>
        <p>If you received this email, your email configuration is working correctly!</p>
    </div>
</body>
</html>`, subject, strings.ReplaceAll(body, "\n", "<br>"), time.Now().Format("2006-01-02 15:04:05 MST"))
	}

	// Record email notification (use the authenticated user's ID)
	notificationID, err := s.recordEmailNotification(userID, to, "test", subject, "test_email", EmailData{
		UserEmail:   to,
		CompanyName: "BOME",
		BaseURL:     s.getBaseURL(),
	})
	if err != nil {
		log.Printf("⚠️ [EMAIL] Failed to record test notification: %v", err)
	}

	// Send with Resend
	sendErr := s.sendWithResend(to, subject, htmlBody, notificationID)
	if sendErr != nil {
		log.Printf("❌ [EMAIL] Resend failed for test email: %v", sendErr)
		return fmt.Errorf("failed to send test email: %w", sendErr)
	}

	log.Printf("✅ [EMAIL] Test email sent successfully to %s", to)
	return nil
}

// SendSubscriptionConfirmation sends a subscription confirmation email
func (s *EmailService) SendSubscriptionConfirmation(userID int, email, name, subscriptionName string, amount float64, currency, periodEnd string) error {
	log.Printf("🔍 [EMAIL] Sending subscription confirmation to: %s", email)

	// Format amount
	amountStr := fmt.Sprintf("%.2f", amount)

	// Prepare email data
	// Get support email from database
	supportEmail, err := s.db.GetEmailSetting("support_email")
	if err != nil || supportEmail == "" {
		supportEmail = "support@bookofmormonevidence.org" // fallback
	}

	emailData := EmailData{
		UserName:         name,
		UserEmail:        email,
		SubscriptionName: subscriptionName,
		Amount:           amountStr,
		Currency:         strings.ToUpper(currency),
		PeriodEnd:        periodEnd,
		CompanyName:      "BOME",
		SupportEmail:     supportEmail,
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

// IsConfigured returns whether the email service is properly configured
func (s *EmailService) IsConfigured() bool {
	// Email service is configured if we have a database connection
	// In the future, this could check SMTP settings
	return s.db != nil
}

// SendWelcomeEmail sends a welcome email (implementing EmailPort)
func (s *EmailService) SendWelcomeEmail(email, firstName string) error {
	// TODO: Implement proper welcome email template
	return nil
}

// SendTemplatedEmail sends an email using a template (implementing EmailPort)
func (s *EmailService) SendTemplatedEmail(to, subject, templateName string, data map[string]interface{}) error {
	// Convert map to EmailData structure
	emailData := EmailData{
		SupportEmail: "support@example.com",
		CompanyName:  "BOME",
		BaseURL:      "https://example.com",
	}

	// Extract common fields from map
	if userName, ok := data["UserName"].(string); ok {
		emailData.UserName = userName
	}
	if userEmail, ok := data["UserEmail"].(string); ok {
		emailData.UserEmail = userEmail
	}
	if verificationURL, ok := data["VerificationURL"].(string); ok {
		emailData.VerificationURL = verificationURL
	}

	// Call the internal sendTemplatedEmail method
	return s.sendTemplatedEmail(to, subject, templateName, emailData, "support@example.com", 0)
}

// IsValidEmail validates an email address format
func (s *EmailService) IsValidEmail(email string) bool {
	// Use the crypto service for validation
	if s.cryptoService == nil {
		return false
	}
	return s.cryptoService.ValidateEmail(email) == nil
}

// GetSMTPSettings retrieves the current SMTP configuration
func (s *EmailService) GetSMTPSettings() (*SMTPSettings, error) {
	// Return default settings or load from database
	// For now, return nil to indicate not configured
	return nil, fmt.Errorf("SMTP settings not configured")
}

// UpdateSMTPSettings updates the SMTP configuration
func (s *EmailService) UpdateSMTPSettings(settings *SMTPSettings) error {
	// Store SMTP settings in database
	// For now, return not implemented
	return fmt.Errorf("SMTP settings update not implemented")
}
