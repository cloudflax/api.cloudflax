package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/cloudflax/api.cloudflax/internal/user"
)

const (
	accessTokenDuration       = 15 * time.Minute
	refreshTokenDuration      = 7 * 24 * time.Hour
	refreshTokenBytes         = 32
	verificationTokenDuration = 24 * time.Hour
)

// Claims holds the JWT payload for access tokens.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair holds the access token and refresh token issued after a successful login or refresh.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ErrInvalidCredentials is returned when login credentials are wrong.
var ErrInvalidCredentials = fmt.Errorf("invalid credentials")

// ErrInvalidVerificationToken is returned when the email verification token is invalid or expired.
var ErrInvalidVerificationToken = fmt.Errorf("invalid verification token")

// ErrEmailAlreadyVerified is returned when the email is already verified.
var ErrEmailAlreadyVerified = fmt.Errorf("email already verified")

// UserRepository is the subset of the user repository the auth service depends on.
type UserRepository interface {
	GetUser(id string) (*user.User, error)
	GetUserByEmail(email string) (*user.User, error)
	Create(u *user.User) error
	Update(u *user.User) error
}

// Service handles authentication business logic.
type Service struct {
	repository     *Repository
	userRepository UserRepository
	jwtSecret      []byte
}

// NewService creates a new auth service.
func NewService(repository *Repository, userRepository UserRepository, jwtSecret string) *Service {
	return &Service{
		repository:     repository,
		userRepository: userRepository,
		jwtSecret:      []byte(jwtSecret),
	}
}

// Register creates a new user with an email/password credential and a pending
// email verification token. The raw verification token is returned so callers
// can hand it to the user (e.g. embed it in a verification link).
func (service *Service) Register(name, email, password string) (*user.User, string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	token := uuid.New().String()
	expiresAt := time.Now().Add(verificationTokenDuration)

	u := &user.User{
		Name:                       name,
		Email:                      normalizedEmail,
		EmailVerificationToken:     &token,
		EmailVerificationExpiresAt: &expiresAt,
	}
	if err := u.SetPassword(password); err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}
	if err := service.userRepository.Create(u); err != nil {
		return nil, "", err
	}

	provider := &UserAuthProvider{
		UserID:            u.ID,
		Provider:          ProviderCredentials,
		ProviderSubjectID: normalizedEmail,
	}
	if err := service.repository.CreateAuthProvider(provider); err != nil {
		return nil, "", fmt.Errorf("create auth provider: %w", err)
	}

	return u, token, nil
}

// VerifyEmail marks the user's email as verified using the token previously issued
// during registration (or after a resend-verification request).
func (service *Service) VerifyEmail(token string) error {
	u, err := service.repository.FindByVerificationToken(token)
	if err != nil {
		return ErrInvalidVerificationToken
	}

	if u.EmailVerificationExpiresAt != nil && time.Now().After(*u.EmailVerificationExpiresAt) {
		return ErrInvalidVerificationToken
	}

	now := time.Now()
	u.EmailVerifiedAt = &now
	u.EmailVerificationToken = nil
	u.EmailVerificationExpiresAt = nil

	return service.userRepository.Update(u)
}

// ResendVerification generates a fresh email verification token for the given email address.
// In production this would trigger an email send; here the token is returned so the caller
// can deliver it (e.g. log it or return it in a dev-only response field).
func (service *Service) ResendVerification(email string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	u, err := service.userRepository.GetUserByEmail(normalizedEmail)
	if err != nil {
		return "", user.ErrNotFound
	}
	if u.IsEmailVerified() {
		return "", ErrEmailAlreadyVerified
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(verificationTokenDuration)
	u.EmailVerificationToken = &token
	u.EmailVerificationExpiresAt = &expiresAt

	if err := service.userRepository.Update(u); err != nil {
		return "", fmt.Errorf("update verification token: %w", err)
	}
	return token, nil
}

// Login verifies credentials and issues a token pair on success.
func (service *Service) Login(email, password string) (*TokenPair, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	u, err := service.userRepository.GetUserByEmail(normalizedEmail)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !u.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}
	return service.generateTokenPair(u)
}

// RefreshTokens validates an existing refresh token, revokes it (rotation),
// and issues a new token pair.
func (service *Service) RefreshTokens(rawRefreshToken string) (*TokenPair, error) {
	tokenHash := hashToken(rawRefreshToken)
	stored, err := service.repository.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if stored.IsRevoked() || stored.IsExpired() {
		return nil, ErrInvalidCredentials
	}

	if err := service.repository.Revoke(stored.ID); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	u, err := service.userRepository.GetUser(stored.UserID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	return service.generateTokenPair(u)
}

// Logout revokes all active refresh tokens for the given user.
func (service *Service) Logout(userID string) error {
	return service.repository.RevokeAllByUserID(userID)
}

// ValidateAccessToken parses and validates a JWT access token.
// Returns the userID and email embedded in the token claims.
// This method satisfies the middleware.TokenValidator interface.
func (service *Service) ValidateAccessToken(tokenString string) (string, string, error) {
	claims, err := service.parseAccessToken(tokenString)
	if err != nil {
		return "", "", err
	}
	return claims.UserID, claims.Email, nil
}

// parseAccessToken parses the JWT and returns the full Claims struct.
func (service *Service) parseAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return service.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return claims, nil
}

// generateTokenPair creates and stores a new access + refresh token pair for the given user.
func (service *Service) generateTokenPair(u *user.User) (*TokenPair, error) {
	expiresAt := time.Now().Add(accessTokenDuration)
	accessToken, err := service.signAccessToken(u, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	rawRefresh, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshRecord := &RefreshToken{
		UserID:    u.ID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(refreshTokenDuration),
	}
	if err := service.repository.Create(refreshRecord); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
	}, nil
}

// signAccessToken builds and signs a JWT for the given user.
func (service *Service) signAccessToken(u *user.User, expiresAt time.Time) (string, error) {
	claims := &Claims{
		UserID: u.ID,
		Email:  u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   u.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(service.jwtSecret)
}

// generateSecureToken creates a cryptographically random hex-encoded token.
func generateSecureToken() (string, error) {
	bytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken returns the SHA-256 hex hash of a raw token string.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
