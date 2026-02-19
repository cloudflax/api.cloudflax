package auth

import (
	"testing"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupServiceTest(t *testing.T) *Service {
	t.Helper()
	require.NoError(t, database.InitForTesting())
	require.NoError(t, database.RunMigrations(&user.User{}, &RefreshToken{}))

	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	return NewService(authRepository, userRepository, testJWTSecret)
}

func seedUser(t *testing.T, name, email, password string) *user.User {
	t.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(t, u.SetPassword(password))
	require.NoError(t, database.DB.Create(u).Error)
	return u
}

func TestService_Login_Success(t *testing.T) {
	service := setupServiceTest(t)
	seedUser(t, "Alice", "alice@example.com", "password123")

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

func TestService_ValidateAccessToken_Valid(t *testing.T) {
	service := setupServiceTest(t)
	seedUser(t, "Carol", "carol@example.com", "password123")

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
	seedUser(t, "Dave", "dave@example.com", "password123")

	pair, err := service.Login("dave@example.com", "password123")
	require.NoError(t, err)

	otherService := NewService(NewRepository(database.DB), user.NewRepository(database.DB), "different-secret")
	_, _, err = otherService.ValidateAccessToken(pair.AccessToken)
	assert.Error(t, err)
}

func TestService_RefreshTokens_Success(t *testing.T) {
	service := setupServiceTest(t)
	seedUser(t, "Eve", "eve@example.com", "password123")

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
	seedUser(t, "Frank", "frank@example.com", "password123")

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

func TestService_Logout_RevokesAllTokens(t *testing.T) {
	service := setupServiceTest(t)
	u := seedUser(t, "Grace", "grace@example.com", "password123")

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
