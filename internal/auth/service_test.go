package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func hashTokenForTest(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func setupServiceTest(t *testing.T) *Service {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}))

	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	return NewService(authRepository, userRepository, testJWTSecret, &noopEmailSender{})
}

func seedUser(t *testing.T, name, email, password string) *user.User {
	t.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(t, u.SetPassword(password))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

// seedVerifiedUser creates a user with email_verified_at set so Login and Refresh succeed.
func seedVerifiedUser(t *testing.T, name, email, password string) *user.User {
	t.Helper()
	now := time.Now()
	u := &user.User{Name: name, Email: email, EmailVerifiedAt: &now}
	require.NoError(t, u.SetPassword(password))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

func TestService_Login_Success(t *testing.T) {
	service := setupServiceTest(t)
	seedVerifiedUser(t, "Alice", "alice@example.com", "password123")

	pair, err := service.Login("alice@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.False(t, pair.ExpiresAt.IsZero())
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	service := setupServiceTest(t)
	seedUser(t, "Bob", "bob@example.com", "correctpass")

	_, err := service.Login("bob@example.com", "wrongpass")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestService_Login_UnknownEmail(t *testing.T) {
	service := setupServiceTest(t)

	_, err := service.Login("ghost@example.com", "password123")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestService_Login_EmailNotVerified(t *testing.T) {
	service := setupServiceTest(t)
	seedUser(t, "Unverified", "unverified@example.com", "password123")

	_, err := service.Login("unverified@example.com", "password123")
	assert.ErrorIs(t, err, ErrEmailNotVerified)
}

func TestService_ValidateAccessToken_Valid(t *testing.T) {
	service := setupServiceTest(t)
	seedVerifiedUser(t, "Carol", "carol@example.com", "password123")

	pair, err := service.Login("carol@example.com", "password123")
	require.NoError(t, err)

	userID, email, err := service.ValidateAccessToken(pair.AccessToken)
	require.NoError(t, err)
	assert.NotEmpty(t, userID)
	assert.Equal(t, "carol@example.com", email)
}

func TestService_ValidateAccessToken_Invalid(t *testing.T) {
	service := setupServiceTest(t)

	_, _, err := service.ValidateAccessToken("invalid.token.here")
	assert.Error(t, err)
}

func TestService_ValidateAccessToken_WrongSecret(t *testing.T) {
	service := setupServiceTest(t)
	seedVerifiedUser(t, "Dave", "dave@example.com", "password123")

	pair, err := service.Login("dave@example.com", "password123")
	require.NoError(t, err)

	otherService := NewService(NewRepository(database.DB), user.NewRepository(database.DB), "different-secret", &noopEmailSender{})
	_, _, err = otherService.ValidateAccessToken(pair.AccessToken)
	assert.Error(t, err)
}

func TestService_RefreshTokens_Success(t *testing.T) {
	service := setupServiceTest(t)
	seedVerifiedUser(t, "Eve", "eve@example.com", "password123")

	pair, err := service.Login("eve@example.com", "password123")
	require.NoError(t, err)

	newPair, err := service.RefreshTokens(pair.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newPair.AccessToken)
	assert.NotEmpty(t, newPair.RefreshToken)
	assert.NotEqual(t, pair.RefreshToken, newPair.RefreshToken)
}

func TestService_RefreshTokens_Rotation(t *testing.T) {
	service := setupServiceTest(t)
	seedVerifiedUser(t, "Frank", "frank@example.com", "password123")

	pair, err := service.Login("frank@example.com", "password123")
	require.NoError(t, err)

	_, err = service.RefreshTokens(pair.RefreshToken)
	require.NoError(t, err)

	_, err = service.RefreshTokens(pair.RefreshToken)
	assert.ErrorIs(t, err, ErrInvalidCredentials, "reusing a rotated refresh token must fail")
}

func TestService_RefreshTokens_InvalidToken(t *testing.T) {
	service := setupServiceTest(t)

	_, err := service.RefreshTokens("random-invalid-token")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestService_RefreshTokens_EmailNotVerified(t *testing.T) {
	service := setupServiceTest(t)
	u := seedUser(t, "Unverified", "unverified@example.com", "password123")

	// Manually create a refresh token for the unverified user (e.g. from before we required verification).
	rawToken := "test-refresh-token-for-unverified-user"
	hash := hashTokenForTest(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)
	require.NoError(t, service.repository.Create(&RefreshToken{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	}))

	_, err := service.RefreshTokens(rawToken)
	assert.ErrorIs(t, err, ErrEmailNotVerified)
}

func TestService_Register_Success(t *testing.T) {
	service := setupServiceTest(t)

	u, token, err := service.Register("Alice", "alice@example.com", "password123")
	require.NoError(t, err)
	assert.NotEmpty(t, u.ID)
	assert.Equal(t, "alice@example.com", u.Email)
	assert.Nil(t, u.EmailVerifiedAt)
	assert.NotEmpty(t, token)

	var provider UserAuthProvider
	require.NoError(t, database.DB.Where("user_id = ? AND provider = ?", u.ID, ProviderCredentials).First(&provider).Error)
	assert.Equal(t, "alice@example.com", provider.ProviderSubjectID)
}

func TestService_Register_DuplicateEmail(t *testing.T) {
	service := setupServiceTest(t)

	_, _, err := service.Register("Alice", "dup@example.com", "password123")
	require.NoError(t, err)

	_, _, err = service.Register("Bob", "dup@example.com", "password456")
	assert.ErrorIs(t, err, user.ErrDuplicateEmail)
}

func TestService_VerifyEmail_Success(t *testing.T) {
	service := setupServiceTest(t)

	_, token, err := service.Register("Carol", "carol@example.com", "password123")
	require.NoError(t, err)

	require.NoError(t, service.VerifyEmail(token))

	var u user.User
	require.NoError(t, database.DB.Where("email = ?", "carol@example.com").First(&u).Error)
	assert.NotNil(t, u.EmailVerifiedAt)
	assert.Nil(t, u.EmailVerificationToken)
}

func TestService_VerifyEmail_InvalidToken(t *testing.T) {
	service := setupServiceTest(t)

	err := service.VerifyEmail("00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(t, err, ErrInvalidVerificationToken)
}

func TestService_ResendVerification_Success(t *testing.T) {
	service := setupServiceTest(t)

	_, oldToken, err := service.Register("Dan", "dan@example.com", "password123")
	require.NoError(t, err)

	newToken, err := service.ResendVerification("dan@example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, oldToken, newToken)
}

func TestService_ResendVerification_AlreadyVerified(t *testing.T) {
	service := setupServiceTest(t)

	_, token, err := service.Register("Eve", "eve@example.com", "password123")
	require.NoError(t, err)
	require.NoError(t, service.VerifyEmail(token))

	_, err = service.ResendVerification("eve@example.com")
	assert.ErrorIs(t, err, ErrEmailAlreadyVerified)
}

func TestService_Logout_RevokesAllTokens(t *testing.T) {
	service := setupServiceTest(t)
	u := seedVerifiedUser(t, "Grace", "grace@example.com", "password123")

	pair1, err := service.Login("grace@example.com", "password123")
	require.NoError(t, err)
	pair2, err := service.Login("grace@example.com", "password123")
	require.NoError(t, err)

	require.NoError(t, service.Logout(u.ID))

	_, err = service.RefreshTokens(pair1.RefreshToken)
	assert.ErrorIs(t, err, ErrInvalidCredentials, "first session token should be revoked after logout")

	_, err = service.RefreshTokens(pair2.RefreshToken)
	assert.ErrorIs(t, err, ErrInvalidCredentials, "second session token should be revoked after logout")
}
