package usecases

import (
	"fmt"
	"time"

	authModels "bome-backend/authentication/models"
	"bome-backend/infrastructure/database"
	"bome-backend/services/security/crypto"
)

// LoginUserInput represents the input for user login
type LoginUserInput struct {
	Email      string
	Password   string
	ClientIP   string
	DeviceInfo string
	UserAgent  string
}

// LoginUserOutput represents the output of user login
type LoginUserOutput struct {
	User         *authModels.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
	Message      string
}

// LoginUser handles the business logic for user login
type LoginUser struct {
	db        *database.DB
	cryptoSvc *crypto.CryptoService
}

// NewLoginUser creates a new LoginUser use case
func NewLoginUser(db *database.DB, cryptoSvc *crypto.CryptoService) *LoginUser {
	return &LoginUser{
		db:        db,
		cryptoSvc: cryptoSvc,
	}
}

// Execute performs the user login business logic
func (uc *LoginUser) Execute(input LoginUserInput) (*LoginUserOutput, error) {
	// 1. Validate email format
	if err := uc.cryptoSvc.ValidateEmail(input.Email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	// 2. Get user from database
	user, err := authModels.GetUserByEmail(uc.db, input.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 3. Verify password
	if err := uc.cryptoSvc.CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// 4. Check if user is active (IsActive is sql.NullBool: invalid/NULL or false = inactive)
	if !user.IsActive.Valid || !user.IsActive.Bool {
		return nil, fmt.Errorf("account is deactivated. Please contact support.")
	}

	// 5. Check email verification
	if !user.EmailVerified {
		return nil, fmt.Errorf("please verify your email before logging in")
	}

	// 6. Check session limit
	canLogin, err := authModels.CheckSessionLimit(uc.db, user.ID, user.MaxSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to check session limit: %w", err)
	}
	if !canLogin {
		return nil, fmt.Errorf("maximum number of active sessions reached. Please log out from another device.")
	}

	// 7. Generate JWT tokens
	tokenPair, err := uc.cryptoSvc.GenerateTokenPair(user.ID, user.Email, user.Role, user.EmailVerified)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// 8. Create session record
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // Refresh token expiry
	_, err = authModels.CreateSession(uc.db, user.ID, extractTokenID(tokenPair.AccessToken, uc.cryptoSvc), input.DeviceInfo, input.ClientIP, input.UserAgent, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// 9. Update last login timestamp
	if err := authModels.UpdateLastLogin(uc.db, user.ID); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}

	// 10. Create audit log
	loginDesc := "User logged in successfully"
	auditLog := &authModels.AuditLog{
		UserID:     &user.ID,
		Action:     "login",
		IPAddress:  input.ClientIP,
		UserAgent:  input.UserAgent,
		Status:     "success",
		Details:    &loginDesc,
	}

	if err := authModels.CreateAuditLog(uc.db, auditLog); err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return &LoginUserOutput{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		Message:      "Login successful",
	}, nil
}

// extractTokenID extracts the token ID from a JWT token
func extractTokenID(token string, cryptoSvc *crypto.CryptoService) string {
	claims, err := cryptoSvc.ParseToken(token)
	if err != nil {
		return ""
	}
	return claims.TokenID
}
