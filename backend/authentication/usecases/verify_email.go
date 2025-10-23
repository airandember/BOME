package usecases

import (
	"fmt"

	authModels "bome-backend/authentication/models"
	"bome-backend/infrastructure/database"
	"bome-backend/services/security/crypto"
)

// VerifyEmailInput represents the input for email verification
type VerifyEmailInput struct {
	Token  string
	UserID int // Optional, can be extracted from token
}

// VerifyEmailOutput represents the output of email verification
type VerifyEmailOutput struct {
	User    *authModels.User
	Message string
}

// VerifyEmail handles the business logic for email verification
type VerifyEmail struct {
	db        *database.DB
	cryptoSvc *crypto.CryptoService
}

// NewVerifyEmail creates a new VerifyEmail use case
func NewVerifyEmail(db *database.DB, cryptoSvc *crypto.CryptoService) *VerifyEmail {
	return &VerifyEmail{
		db:        db,
		cryptoSvc: cryptoSvc,
	}
}

// Execute performs the email verification business logic
func (uc *VerifyEmail) Execute(input VerifyEmailInput) (*VerifyEmailOutput, error) {
	// 1. Validate token format
	if input.Token == "" {
		return nil, fmt.Errorf("verification token is required")
	}

	// 2. Get user by verification token
	user, err := authModels.GetUserByVerificationToken(uc.db, input.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired verification token")
	}

	if user == nil {
		return nil, fmt.Errorf("invalid or expired verification token")
	}

	// 3. Check if email is already verified
	if user.EmailVerified {
		return &VerifyEmailOutput{
			User:    user,
			Message: "Email is already verified",
		}, nil
	}

	// 4. Verify email
	if err := authModels.SetUserEmailVerified(uc.db, user.ID, true); err != nil {
		return nil, fmt.Errorf("failed to verify email: %w", err)
	}

	// 5. Clear verification token
	if err := authModels.ClearVerificationToken(uc.db, user.ID); err != nil {
		// Log error but don't fail verification
		fmt.Printf("Failed to clear verification token: %v\n", err)
	}

	// 6. Create audit log
	auditLog := &authModels.AuditLog{
		UserID:      user.ID,
		Action:      "email_verified",
		IPAddress:   "", // Not available in use case
		Status:      "success",
		Description: "Email verified successfully",
	}

	if err := authModels.CreateAuditLog(uc.db, auditLog); err != nil {
		// Log error but don't fail verification
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	// Update user object
	user.EmailVerified = true

	return &VerifyEmailOutput{
		User:    user,
		Message: "Email verified successfully",
	}, nil
}
