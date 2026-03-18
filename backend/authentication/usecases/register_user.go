package usecases

import (
	"fmt"
	"strings"

	authModels "bome-backend/authentication/models"
	"bome-backend/infrastructure/database"
	"bome-backend/services/communication/email"
	"bome-backend/services/security/crypto"
)

// RegisterUserInput represents the input for user registration
type RegisterUserInput struct {
	Email     string
	FirstName string
	LastName  string
}

// RegisterUserOutput represents the output of user registration
type RegisterUserOutput struct {
	User    *authModels.User
	Message string
}

// RegisterUser handles the business logic for user registration
type RegisterUser struct {
	db           *database.DB
	cryptoSvc    *crypto.CryptoService
	emailService *email.EmailService
}

// NewRegisterUser creates a new RegisterUser use case
func NewRegisterUser(db *database.DB, cryptoSvc *crypto.CryptoService, emailService *email.EmailService) *RegisterUser {
	return &RegisterUser{
		db:           db,
		cryptoSvc:    cryptoSvc,
		emailService: emailService,
	}
}

// Execute performs the user registration business logic
func (uc *RegisterUser) Execute(input RegisterUserInput) (*RegisterUserOutput, error) {
	// 1. Validate and sanitize input
	input.Email = strings.ToLower(uc.cryptoSvc.SanitizeString(input.Email))
	input.FirstName = uc.cryptoSvc.SanitizeString(input.FirstName)
	input.LastName = uc.cryptoSvc.SanitizeString(input.LastName)

	// 2. Validate email format
	if err := uc.cryptoSvc.ValidateEmail(input.Email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	// 3. Validate names
	if err := uc.cryptoSvc.ValidateName(input.FirstName); err != nil {
		return nil, fmt.Errorf("invalid first name: %w", err)
	}

	if err := uc.cryptoSvc.ValidateName(input.LastName); err != nil {
		return nil, fmt.Errorf("invalid last name: %w", err)
	}

	// 4. Check if user already exists
	existingUser, err := authModels.GetUserByEmail(uc.db, input.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", input.Email)
	}

	// 5. Create user with unverified email (temporary password until they set one after verification)
	tempPassword := uc.cryptoSvc.GenerateSecureToken()
	passwordHash, err := uc.cryptoSvc.HashPassword(tempPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Save user to database
	newUser, err := authModels.CreateUser(uc.db, input.Email, passwordHash, input.FirstName, input.LastName, "user")
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 7. Send verification email
	if uc.emailService != nil {
		// Generate verification token
		token := uc.cryptoSvc.GenerateSecureToken()

		// Store token in database
		if err := authModels.SetVerificationToken(uc.db, newUser.ID, token); err != nil {
			// Log error but don't fail registration
			fmt.Printf("Failed to store verification token: %v\n", err)
		} else {
			// Send verification email
			if err := uc.emailService.SendVerificationEmail(newUser.ID, input.Email, input.FirstName); err != nil {
				// Log error but don't fail registration
				fmt.Printf("Failed to send verification email: %v\n", err)
			}
		}
	}

	return &RegisterUserOutput{
		User:    newUser,
		Message: "User registered successfully. Please check your email to verify your account.",
	}, nil
}
