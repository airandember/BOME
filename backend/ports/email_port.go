package ports

// EmailPort defines the interface for email operations
type EmailPort interface {
	// Email Sending
	SendVerificationEmail(email, token, firstName string) error
	SendPasswordResetEmail(email, token, firstName string) error
	SendWelcomeEmail(email, firstName string) error
	SendPasswordSetupEmail(email, token, firstName string) error

	// Template Email Sending
	SendTemplatedEmail(to, subject, templateName string, data map[string]interface{}) error

	// Email Validation
	IsValidEmail(email string) bool

	// Service Status
	IsConfigured() bool
}

// EmailMessage represents an email message structure
type EmailMessage struct {
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	From        string
	ReplyTo     string
	Attachments []EmailAttachment
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename string
	Content  []byte
	MimeType string
}
