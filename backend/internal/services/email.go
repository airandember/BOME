package services

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService handles all email operations via SendGrid
type EmailService struct {
	client      *sendgrid.Client
	fromEmail   string
	fromName    string
	templateIDs map[string]string
	baseURL     string
}

// EmailData represents data for email templates
type EmailData struct {
	User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"user"`
	Subject    string                 `json:"subject"`
	Content    string                 `json:"content"`
	ActionURL  string                 `json:"action_url"`
	ExpiresAt  time.Time              `json:"expires_at"`
	CustomData map[string]interface{} `json:"custom_data"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

// NewEmailService creates a new SendGrid email service instance
func NewEmailService() *EmailService {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")
	fromName := os.Getenv("SENDGRID_FROM_NAME")
	baseURL := os.Getenv("PUBLIC_APP_URL")

	// Initialize template IDs
	templateIDs := map[string]string{
		"welcome":            os.Getenv("SENDGRID_TEMPLATE_ID_WELCOME"),
		"password_reset":     os.Getenv("SENDGRID_TEMPLATE_ID_PASSWORD_RESET"),
		"subscription":       os.Getenv("SENDGRID_TEMPLATE_ID_SUBSCRIPTION"),
		"email_verification": os.Getenv("SENDGRID_TEMPLATE_ID_EMAIL_VERIFICATION"),
	}

	return &EmailService{
		client:      sendgrid.NewSendClient(apiKey),
		fromEmail:   fromEmail,
		fromName:    fromName,
		templateIDs: templateIDs,
		baseURL:     baseURL,
	}
}

// SendEmail sends a simple email
func (e *EmailService) SendEmail(to, subject, htmlContent, textContent string) error {
	from := mail.NewEmail(e.fromName, e.fromEmail)
	toEmail := mail.NewEmail("", to)

	message := mail.NewSingleEmail(from, subject, toEmail, textContent, htmlContent)

	response, err := e.client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("email send failed with status %d: %s", response.StatusCode, response.Body)
	}

	return nil
}

// SendTemplateEmail sends an email using a SendGrid template
func (e *EmailService) SendTemplateEmail(to, templateName string, data EmailData) error {
	// For now, send a simple email with the data
	subject := data.Subject
	content := data.Content

	if data.ActionURL != "" {
		content += "\n\nAction URL: " + data.ActionURL
	}

	return e.SendEmail(to, subject, content, content)
}

// SendWelcomeEmail sends a welcome email to new users
func (e *EmailService) SendWelcomeEmail(name, email string) error {
	data := EmailData{
		Subject: "Welcome to Book of Mormon Evidences!",
		Content: "Thank you for joining our community. We're excited to share compelling evidence with you.",
	}
	data.User.Name = name
	data.User.Email = email
	data.ActionURL = fmt.Sprintf("%s/dashboard", e.baseURL)

	return e.SendTemplateEmail(email, "welcome", data)
}

// SendPasswordResetEmail sends a password reset email
func (e *EmailService) SendPasswordResetEmail(name, email, resetToken string) error {
	data := EmailData{
		Subject:   "Reset Your Password",
		Content:   "Click the link below to reset your password. This link will expire in 1 hour.",
		ActionURL: fmt.Sprintf("%s/reset-password?token=%s", e.baseURL, resetToken),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	data.User.Name = name
	data.User.Email = email

	return e.SendTemplateEmail(email, "password_reset", data)
}

// SendEmailVerification sends an email verification email
func (e *EmailService) SendEmailVerification(name, email, verificationToken string) error {
	data := EmailData{
		Subject:   "Verify Your Email Address",
		Content:   "Please verify your email address by clicking the link below.",
		ActionURL: fmt.Sprintf("%s/verify-email?token=%s", e.baseURL, verificationToken),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	data.User.Name = name
	data.User.Email = email

	return e.SendTemplateEmail(email, "email_verification", data)
}

// SendSubscriptionConfirmation sends a subscription confirmation email
func (e *EmailService) SendSubscriptionConfirmation(name, email, planName string, amount float64) error {
	data := EmailData{
		Subject:   "Subscription Confirmed",
		Content:   fmt.Sprintf("Your %s subscription has been activated. Thank you for your support!", planName),
		ActionURL: fmt.Sprintf("%s/account", e.baseURL),
		CustomData: map[string]interface{}{
			"plan_name": planName,
			"amount":    amount,
			"currency":  "USD",
		},
	}
	data.User.Name = name
	data.User.Email = email

	return e.SendTemplateEmail(email, "subscription", data)
}

// SendSubscriptionCancellation sends a subscription cancellation email
func (e *EmailService) SendSubscriptionCancellation(name, email, planName string) error {
	data := EmailData{
		Subject:   "Subscription Cancelled",
		Content:   fmt.Sprintf("Your %s subscription has been cancelled. You'll continue to have access until the end of your billing period.", planName),
		ActionURL: fmt.Sprintf("%s/account", e.baseURL),
		CustomData: map[string]interface{}{
			"plan_name": planName,
		},
	}
	data.User.Name = name
	data.User.Email = email

	return e.SendTemplateEmail(email, "subscription", data)
}

// SendPaymentFailed sends a payment failure notification
func (e *EmailService) SendPaymentFailed(name, email, planName string) error {
	data := EmailData{
		Subject:   "Payment Failed",
		Content:   "We were unable to process your payment. Please update your payment method to continue your subscription.",
		ActionURL: fmt.Sprintf("%s/billing", e.baseURL),
		CustomData: map[string]interface{}{
			"plan_name": planName,
		},
	}
	data.User.Name = name
	data.User.Email = email

	return e.SendTemplateEmail(email, "subscription", data)
}

// SendNewVideoNotification sends a notification about new video content
func (e *EmailService) SendNewVideoNotification(name, email, videoTitle, videoURL string) error {
	subject := "New Video Available: " + videoTitle
	content := fmt.Sprintf("A new video '%s' is now available for viewing.", videoTitle)

	data := EmailData{
		Subject:   subject,
		Content:   content,
		ActionURL: videoURL,
		CustomData: map[string]interface{}{
			"video_title": videoTitle,
			"video_url":   videoURL,
		},
	}
	data.User.Name = name
	data.User.Email = email

	return e.SendTemplateEmail(email, "subscription", data)
}

// SendAdminNotification sends a notification to admin users
func (e *EmailService) SendAdminNotification(subject, content string) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		return fmt.Errorf("admin email not configured")
	}

	return e.SendEmail(adminEmail, subject, content, content)
}

// SendBulkEmail sends emails to multiple recipients
func (e *EmailService) SendBulkEmail(recipients []string, subject, htmlContent, textContent string) error {
	from := mail.NewEmail(e.fromName, e.fromEmail)

	// Send individual emails to each recipient
	for _, recipient := range recipients {
		toEmail := mail.NewEmail("", recipient)
		message := mail.NewSingleEmail(from, subject, toEmail, textContent, htmlContent)

		response, err := e.client.Send(message)
		if err != nil {
			return fmt.Errorf("failed to send bulk email to %s: %w", recipient, err)
		}

		if response.StatusCode >= 400 {
			return fmt.Errorf("bulk email send failed to %s with status %d: %s", recipient, response.StatusCode, response.Body)
		}
	}

	return nil
}

// GenerateEmailTemplate generates an email template from HTML and text templates
func (e *EmailService) GenerateEmailTemplate(templateName string, data EmailData) (*EmailTemplate, error) {
	// Load template files
	htmlTemplate, err := template.ParseFiles(fmt.Sprintf("templates/emails/%s.html", templateName))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	textTemplate, err := template.ParseFiles(fmt.Sprintf("templates/emails/%s.txt", templateName))
	if err != nil {
		return nil, fmt.Errorf("failed to parse text template: %w", err)
	}

	// Execute templates
	var htmlBuffer, textBuffer bytes.Buffer

	if err := htmlTemplate.Execute(&htmlBuffer, data); err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}

	if err := textTemplate.Execute(&textBuffer, data); err != nil {
		return nil, fmt.Errorf("failed to execute text template: %w", err)
	}

	return &EmailTemplate{
		Subject: data.Subject,
		HTML:    htmlBuffer.String(),
		Text:    textBuffer.String(),
	}, nil
}

// ValidateEmail validates an email address format
func (e *EmailService) ValidateEmail(email string) bool {
	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	return true
}

// Helper method to convert EmailData to SendGrid template format
func (e *EmailService) convertToTemplateData(data EmailData) map[string]interface{} {
	return map[string]interface{}{
		"user":        data.User,
		"subject":     data.Subject,
		"content":     data.Content,
		"action_url":  data.ActionURL,
		"expires_at":  data.ExpiresAt.Format(time.RFC3339),
		"custom_data": data.CustomData,
		"base_url":    e.baseURL,
	}
}

// GetEmailStats retrieves email sending statistics
func (e *EmailService) GetEmailStats() (map[string]interface{}, error) {
	// This would typically call SendGrid's statistics API
	// For now, return a placeholder
	return map[string]interface{}{
		"sent_today":      0,
		"sent_this_week":  0,
		"sent_this_month": 0,
		"bounce_rate":     0.0,
		"open_rate":       0.0,
		"click_rate":      0.0,
	}, nil
}
