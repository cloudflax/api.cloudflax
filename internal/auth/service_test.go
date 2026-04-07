package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/shared/verificationnotify"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingNotifier struct{}

func (failingNotifier) NotifyVerificationEmail(context.Context, string, string, string) error {
	return errors.New("notifier failed")
}

type stubPasswordResetNotifier struct {
	lastLink      string
	lastExpiresIn string
	err           error
	calls         int
}

func (s *stubPasswordResetNotifier) NotifyPasswordResetEmail(ctx context.Context, toEmail, name, link, expiresIn string) error {
	s.calls++
	s.lastLink = link
	s.lastExpiresIn = expiresIn
	return s.err
}

func resetTokenFromLink(link string) string {
	const prefix = "token="
	idx := strings.Index(link, prefix)
	if idx < 0 {
		return ""
	}
	return link[idx+len(prefix):]
}

func seedCredentialsProvider(test *testing.T, userID, email string) {
	test.Helper()
	authRepo := NewRepository(database.DB)
	require.NoError(test, authRepo.CreateAuthProvider(&UserAuthProvider{
		UserID:            userID,
		Provider:          ProviderCredentials,
		ProviderSubjectID: strings.ToLower(strings.TrimSpace(email)),
	}))
}

// En: hashTokenForTest calculates the SHA-256 hash of a test token.
// Es: hashTokenForTest calcula el hash SHA-256 de un token de prueba.
func hashTokenForTest(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// En: setupServiceTest sets up the authentication service for tests.
// Es: setupServiceTest configura el servicio de autenticación para pruebas.
func setupServiceTest(test *testing.T) *Service {
	test.Helper()
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))

	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	return NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:            testJWTSecret,
		VerificationNotifier: verificationnotify.NoopNotifier{},
		FrontendURL:          "http://test",
	})
}

// En: seedUser creates a test user.
// Es: seedUser crea un usuario para pruebas.
func seedUser(test *testing.T, name, email, password string) *user.User {
	test.Helper()
	u := &user.User{Name: name, Email: email}
	require.NoError(test, u.SetPassword(password))
	require.NoError(test, database.DB.Create(u).Error)
	return u
}

// En: seedVerifiedUser creates a verified test user.
// Es: seedVerifiedUser crea un usuario con emailverifiedat establecido para que Login y Refresh tengan éxito.
func seedVerifiedUser(test *testing.T, name, email, password string) *user.User {
	test.Helper()
	now := time.Now()
	u := &user.User{Name: name, Email: email, EmailVerifiedAt: &now}
	require.NoError(test, u.SetPassword(password))
	require.NoError(test, database.DB.Create(u).Error)
	return u
}

// En: TestServiceLoginSuccess tests the successful login.
// Es: TestServiceLoginSuccess prueba el inicio de sesión exitoso.
func TestServiceLoginSuccess(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Alice", "alice@example.com", "password123")

	pair, err := service.Login("alice@example.com", "password123")
	require.NoError(test, err)
	assert.NotEmpty(test, pair.AccessToken)
	assert.NotEmpty(test, pair.RefreshToken)
	assert.False(test, pair.ExpiresAt.IsZero())
}

// En: TestServiceLoginInvalidCredentials tests the login with invalid credentials.
// Es: TestServiceLoginInvalidCredentials prueba el inicio de sesión con credenciales inválidas.
func TestServiceLoginInvalidCredentials(test *testing.T) {
	service := setupServiceTest(test)
	seedUser(test, "Bob", "bob@example.com", "correctpass")

	_, err := service.Login("bob@example.com", "wrongpass")
	assert.ErrorIs(test, err, ErrInvalidCredentials)
}

// En: TestServiceLoginUnknownEmail tests the login with an unknown email.
// Es: TestServiceLoginUnknownEmail prueba el inicio de sesión con correo electrónico desconocido.
func TestServiceLoginUnknownEmail(test *testing.T) {
	service := setupServiceTest(test)

	_, err := service.Login("ghost@example.com", "password123")
	assert.ErrorIs(test, err, ErrInvalidCredentials)
}

// En: TestServiceLoginEmailNotVerified tests the login with an unverified email.
// Es: TestServiceLoginEmailNotVerified prueba el inicio de sesión con correo electrónico no verificado.
func TestServiceLoginEmailNotVerified(test *testing.T) {
	service := setupServiceTest(test)
	seedUser(test, "Unverified", "unverified@example.com", "password123")

	_, err := service.Login("unverified@example.com", "password123")
	assert.ErrorIs(test, err, ErrEmailNotVerified)
}

// En: TestServiceLoginCredentialLockoutAfterMaxFailures locks the email after repeated wrong passwords (B3).
// Es: TestServiceLoginCredentialLockoutAfterMaxFailures bloquea el email tras contraseñas incorrectas repetidas (B3).
func TestServiceLoginCredentialLockoutAfterMaxFailures(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Lock", "lockme@example.com", "rightpass")

	for range loginCredentialLockoutMaxFailures - 1 {
		_, err := service.Login("lockme@example.com", "wrong")
		require.ErrorIs(test, err, ErrInvalidCredentials)
	}
	_, err := service.Login("lockme@example.com", "rightpass")
	require.NoError(test, err)

	for range loginCredentialLockoutMaxFailures {
		_, err := service.Login("lockme@example.com", "wrong")
		require.ErrorIs(test, err, ErrInvalidCredentials)
	}

	_, err = service.Login("lockme@example.com", "rightpass")
	var lockErr *CredentialsLockedError
	require.ErrorAs(test, err, &lockErr)
	assert.Greater(test, lockErr.RetryAfter, time.Duration(0))
	assert.LessOrEqual(test, lockErr.RetryAfter, loginCredentialLockoutDuration+time.Second)
}

// En: TestServiceLoginCredentialLockoutUnknownEmail applies the same counters to non-existent addresses.
// Es: TestServiceLoginCredentialLockoutUnknownEmail aplica los mismos contadores a direcciones inexistentes.
func TestServiceLoginCredentialLockoutUnknownEmail(test *testing.T) {
	service := setupServiceTest(test)
	email := "nouser@example.com"

	for range loginCredentialLockoutMaxFailures {
		_, err := service.Login(email, "any")
		require.ErrorIs(test, err, ErrInvalidCredentials)
	}
	_, err := service.Login(email, "any")
	var lockUnknown *CredentialsLockedError
	require.ErrorAs(test, err, &lockUnknown)
	assert.Greater(test, lockUnknown.RetryAfter, time.Duration(0))
}

// En: TestServiceLoginClearsCredentialLockoutOnSuccess removes counters after a valid login.
// Es: TestServiceLoginClearsCredentialLockoutOnSuccess elimina contadores tras un login valido.
func TestServiceLoginClearsCredentialLockoutOnSuccess(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Clear", "clear@example.com", "password123")

	for range 3 {
		_, err := service.Login("clear@example.com", "wrong")
		require.ErrorIs(test, err, ErrInvalidCredentials)
	}

	pair, err := service.Login("clear@example.com", "password123")
	require.NoError(test, err)
	require.NotNil(test, pair)

	var count int64
	require.NoError(test, database.DB.Model(&LoginCredentialLockout{}).Where("email_normalized = ?", "clear@example.com").Count(&count).Error)
	assert.Equal(test, int64(0), count)
}

// En: TestServiceResetPasswordClearsCredentialLockout resets lockout state after a successful password reset.
// Es: TestServiceResetPasswordClearsCredentialLockout reinicia el bloqueo tras un reset de contraseña exitoso.
func TestServiceResetPasswordClearsCredentialLockout(test *testing.T) {
	service := setupServiceTest(test)
	u := seedVerifiedUser(test, "ResetLock", "resetlock@example.com", "oldpass")

	lockUntil := time.Now().Add(loginCredentialLockoutDuration)
	require.NoError(test, database.DB.Create(&LoginCredentialLockout{
		EmailNormalized: u.Email,
		FailedCount:     loginCredentialLockoutMaxFailures,
		WindowStart:     time.Now(),
		LockedUntil:     &lockUntil,
	}).Error)

	rawToken := "reset-clear-lockout-token"
	require.NoError(test, database.DB.Model(u).Updates(map[string]any{
		"password_reset_token_hash": hashTokenForTest(rawToken),
		"password_reset_expires_at": time.Now().Add(time.Hour),
	}).Error)

	require.NoError(test, service.ResetPassword(rawToken, "newpass12345"))

	var count int64
	require.NoError(test, database.DB.Model(&LoginCredentialLockout{}).Where("email_normalized = ?", u.Email).Count(&count).Error)
	assert.Equal(test, int64(0), count)
}

// En: TestServicePeekEmailVerificationToken tests peek without resend side effects.
// Es: TestServicePeekEmailVerificationToken prueba la lectura del token sin efectos de reenvío.
func TestServicePeekEmailVerificationToken(test *testing.T) {
	service := setupServiceTest(test)
	_, tokenFromRegister, err := service.Register("Peek User", "peek@example.com", "password123")
	require.NoError(test, err)
	require.NotEmpty(test, tokenFromRegister)

	got, err := service.PeekEmailVerificationToken("peek@example.com")
	require.NoError(test, err)
	assert.Equal(test, tokenFromRegister, got)

	_, err = service.PeekEmailVerificationToken("missing@example.com")
	assert.ErrorIs(test, err, user.ErrNotFound)
}

// En: TestServiceValidateAccessTokenValid tests the validation of a valid access token.
// Es: TestServiceValidateAccessTokenValid prueba la validación de token de acceso válido.
func TestServiceValidateAccessTokenValid(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Carol", "carol@example.com", "password123")

	pair, err := service.Login("carol@example.com", "password123")
	require.NoError(test, err)

	userID, email, err := service.ValidateAccessToken(pair.AccessToken)
	require.NoError(test, err)
	assert.NotEmpty(test, userID)
	assert.Equal(test, "carol@example.com", email)
}

// En: TestServiceValidateAccessTokenInvalid tests the validation of an invalid access token.
// Es: TestServiceValidateAccessTokenInvalid prueba la validación de token de acceso inválido.
func TestServiceValidateAccessTokenInvalid(test *testing.T) {
	service := setupServiceTest(test)

	_, _, err := service.ValidateAccessToken("invalid.token.here")
	assert.Error(test, err)
}

// En: TestServiceValidateAccessTokenWrongSecret tests the validation of an access token with a wrong secret.
// Es: TestServiceValidateAccessTokenWrongSecret prueba la validación de token de acceso con secreto incorrecto.
func TestServiceValidateAccessTokenWrongSecret(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Dave", "dave@example.com", "password123")

	pair, err := service.Login("dave@example.com", "password123")
	require.NoError(test, err)

	otherService := NewService(NewRepository(database.DB), user.NewRepository(database.DB), ServiceOptions{
		JWTSecret:            "different-secret",
		VerificationNotifier: verificationnotify.NoopNotifier{},
		FrontendURL:          "http://test",
	})
	_, _, err = otherService.ValidateAccessToken(pair.AccessToken)
	assert.Error(test, err)
}

// En: TestServiceRefreshTokensSuccess tests the successful refresh of tokens.
// Es: TestServiceRefreshTokensSuccess prueba el refresco de tokens exitoso.
func TestServiceRefreshTokensSuccess(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Eve", "eve@example.com", "password123")

	pair, err := service.Login("eve@example.com", "password123")
	require.NoError(test, err)

	newPair, err := service.RefreshTokens(pair.RefreshToken)
	require.NoError(test, err)
	assert.NotEmpty(test, newPair.AccessToken)
	assert.NotEmpty(test, newPair.RefreshToken)
	assert.NotEqual(test, pair.RefreshToken, newPair.RefreshToken)
}

// En: TestServiceRefreshTokensRotation tests the refresh of tokens with rotation.
// Es: TestServiceRefreshTokensRotation prueba el refresco de tokens con rotación.
func TestServiceRefreshTokensRotation(test *testing.T) {
	service := setupServiceTest(test)
	seedVerifiedUser(test, "Frank", "frank@example.com", "password123")

	pair, err := service.Login("frank@example.com", "password123")
	require.NoError(test, err)

	_, err = service.RefreshTokens(pair.RefreshToken)
	require.NoError(test, err)

	_, err = service.RefreshTokens(pair.RefreshToken)
	assert.ErrorIs(test, err, ErrInvalidCredentials, "reusing a rotated refresh token must fail")
}

// En: TestServiceRefreshTokensInvalidToken tests the refresh of tokens with an invalid token.
// Es: TestServiceRefreshTokensInvalidToken prueba el refresco de tokens con token inválido.
func TestServiceRefreshTokensInvalidToken(test *testing.T) {
	service := setupServiceTest(test)

	_, err := service.RefreshTokens("random-invalid-token")
	assert.ErrorIs(test, err, ErrInvalidCredentials)
}

// En: TestServiceRefreshTokensEmailNotVerified tests the refresh of tokens with an unverified email.
// Es: TestServiceRefreshTokensEmailNotVerified prueba el refresco de tokens con correo electrónico no verificado.
func TestServiceRefreshTokensEmailNotVerified(test *testing.T) {
	service := setupServiceTest(test)
	u := seedUser(test, "Unverified", "unverified@example.com", "password123")

	// Manually create a refresh token for the unverified user (e.g. from before we required verification).
	rawToken := "test-refresh-token-for-unverified-user"
	hash := hashTokenForTest(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)
	require.NoError(test, service.repository.Create(&RefreshToken{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	}))

	_, err := service.RefreshTokens(rawToken)
	assert.ErrorIs(test, err, ErrEmailNotVerified)
}

// En: TestServiceRegisterSuccess tests the successful registration.
// Es: TestServiceRegisterSuccess prueba el registro exitoso.
func TestServiceRegisterSuccess(test *testing.T) {
	service := setupServiceTest(test)

	u, token, err := service.Register("Alice", "alice@example.com", "password123")
	require.NoError(test, err)
	assert.NotEmpty(test, u.ID)
	assert.Equal(test, "alice@example.com", u.Email)
	assert.Nil(test, u.EmailVerifiedAt)
	assert.NotEmpty(test, token)

	var provider UserAuthProvider
	require.NoError(test, database.DB.Where("user_id = ? AND provider = ?", u.ID, ProviderCredentials).First(&provider).Error)
	assert.Equal(test, "alice@example.com", provider.ProviderSubjectID)
}

// En: TestServiceRegisterDuplicateEmail tests the registration with a duplicate email.
// Es: TestServiceRegisterDuplicateEmail prueba el registro con correo electrónico duplicado.
func TestServiceRegisterDuplicateEmail(test *testing.T) {
	service := setupServiceTest(test)

	_, _, err := service.Register("Alice", "dup@example.com", "password123")
	require.NoError(test, err)

	_, _, err = service.Register("Bob", "dup@example.com", "password456")
	assert.ErrorIs(test, err, user.ErrDuplicateEmail)
}

// En: TestServiceVerifyEmailSuccess tests the successful email verification.
// Es: TestServiceVerifyEmailSuccess prueba la verificación de correo electrónico exitosa.
func TestServiceVerifyEmailSuccess(test *testing.T) {
	service := setupServiceTest(test)

	_, token, err := service.Register("Carol", "carol@example.com", "password123")
	require.NoError(test, err)

	require.NoError(test, service.VerifyEmail(token))

	var u user.User
	require.NoError(test, database.DB.Where("email = ?", "carol@example.com").First(&u).Error)
	assert.NotNil(test, u.EmailVerifiedAt)
	assert.Nil(test, u.EmailVerificationToken)
}

// En: TestServiceVerifyEmailInvalidToken tests the email verification with an invalid token.
// Es: TestServiceVerifyEmailInvalidToken prueba la verificación de correo electrónico con token inválido.
func TestServiceVerifyEmailInvalidToken(test *testing.T) {
	service := setupServiceTest(test)

	err := service.VerifyEmail("00000000-0000-0000-0000-000000000000")
	assert.ErrorIs(test, err, ErrInvalidVerificationToken)
}

// En: TestServiceResendVerificationSuccess tests the successful email verification resend.
// Es: TestServiceResendVerificationSuccess prueba el envío de correo de verificación exitoso.
func TestServiceResendVerificationSuccess(test *testing.T) {
	service := setupServiceTest(test)

	_, oldToken, err := service.Register("Dan", "dan@example.com", "password123")
	require.NoError(test, err)

	newToken, err := service.ResendVerification("dan@example.com")
	require.NoError(test, err)
	assert.NotEmpty(test, newToken)
	assert.NotEqual(test, oldToken, newToken)
}

// En: TestServiceResendVerificationAlreadyVerified tests the email verification resend with an already verified email.
// Es: TestServiceResendVerificationAlreadyVerified prueba el envío de correo de verificación con correo electrónico ya verificado.
func TestServiceResendVerificationAlreadyVerified(test *testing.T) {
	service := setupServiceTest(test)

	_, token, err := service.Register("Eve", "eve@example.com", "password123")
	require.NoError(test, err)
	require.NoError(test, service.VerifyEmail(token))

	_, err = service.ResendVerification("eve@example.com")
	assert.ErrorIs(test, err, ErrEmailAlreadyVerified)
}

// En: TestServiceResendVerificationEmailSendFailure returns error on notifier failure.
// Es: TestServiceResendVerificationEmailSendFailure devuelve error si falla notifier.
func TestServiceResendVerificationEmailSendFailure(test *testing.T) {
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))

	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:            testJWTSecret,
		VerificationNotifier: failingNotifier{},
		FrontendURL:          "http://test",
	})

	_, _, err := service.Register("Mailer", "mailer@example.com", "password123")
	require.NoError(test, err)

	_, err = service.ResendVerification("mailer@example.com")
	require.Error(test, err)
	assert.Contains(test, err.Error(), "send verification email after resend")
}

// En: TestServiceLogoutRevokesAllTokens tests the logout and revocation of all tokens.
// Es: TestServiceLogoutRevokesAllTokens prueba el cierre de sesión y revocación de todos los tokens.
func TestServiceLogoutRevokesAllTokens(test *testing.T) {
	service := setupServiceTest(test)
	u := seedVerifiedUser(test, "Grace", "grace@example.com", "password123")

	pair1, err := service.Login("grace@example.com", "password123")
	require.NoError(test, err)
	pair2, err := service.Login("grace@example.com", "password123")
	require.NoError(test, err)

	require.NoError(test, service.Logout(u.ID))

	_, err = service.RefreshTokens(pair1.RefreshToken)
	assert.ErrorIs(test, err, ErrInvalidCredentials, "first session token should be revoked after logout")

	_, err = service.RefreshTokens(pair2.RefreshToken)
	assert.ErrorIs(test, err, ErrInvalidCredentials, "second session token should be revoked after logout")
}

// En: TestServiceForgotPasswordUnknownEmail returns nil without notifying.
// Es: TestServiceForgotPasswordUnknownEmail devuelve nil sin notificar.
func TestServiceForgotPasswordUnknownEmail(test *testing.T) {
	stub := &stubPasswordResetNotifier{}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})

	require.NoError(test, service.ForgotPassword(context.Background(), "nobody@example.com"))
	assert.Equal(test, 0, stub.calls)
}

// En: TestServiceForgotPasswordUnverifiedUser does not send email.
// Es: TestServiceForgotPasswordUnverifiedUser no envia correo.
func TestServiceForgotPasswordUnverifiedUser(test *testing.T) {
	stub := &stubPasswordResetNotifier{}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})
	u := seedUser(test, "U", "u@example.com", "password123")
	seedCredentialsProvider(test, u.ID, u.Email)

	require.NoError(test, service.ForgotPassword(context.Background(), "u@example.com"))
	assert.Equal(test, 0, stub.calls)
}

// En: TestServiceForgotPasswordNoCredentialsProvider does not send email.
// Es: TestServiceForgotPasswordNoCredentialsProvider no envia correo.
func TestServiceForgotPasswordNoCredentialsProvider(test *testing.T) {
	stub := &stubPasswordResetNotifier{}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})
	seedVerifiedUser(test, "V", "v@example.com", "password123")

	require.NoError(test, service.ForgotPassword(context.Background(), "v@example.com"))
	assert.Equal(test, 0, stub.calls)
}

// En: TestServiceForgotPasswordSuccess stores token and notifies.
// Es: TestServiceForgotPasswordSuccess guarda token y notifica.
func TestServiceForgotPasswordSuccess(test *testing.T) {
	stub := &stubPasswordResetNotifier{}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})
	u := seedVerifiedUser(test, "W", "w@example.com", "password123")
	seedCredentialsProvider(test, u.ID, u.Email)

	require.NoError(test, service.ForgotPassword(context.Background(), "w@example.com"))
	require.Equal(test, 1, stub.calls)
	assert.Contains(test, stub.lastLink, "http://test/auth/reset-password?token=")
	assert.Equal(test, "60 minutes", stub.lastExpiresIn)

	var dbUser user.User
	require.NoError(test, database.DB.Where("id = ?", u.ID).First(&dbUser).Error)
	require.NotNil(test, dbUser.PasswordResetTokenHash)
	assert.NotEmpty(test, *dbUser.PasswordResetTokenHash)
}

// En: TestServiceForgotPasswordNotifyFailureRollsBackToken clears DB fields when notify fails.
// Es: TestServiceForgotPasswordNotifyFailureRollsBackToken limpia campos en BD si falla el notify.
func TestServiceForgotPasswordNotifyFailureRollsBackToken(test *testing.T) {
	stub := &stubPasswordResetNotifier{err: errors.New("lambda failed")}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})
	u := seedVerifiedUser(test, "X", "x@example.com", "password123")
	seedCredentialsProvider(test, u.ID, u.Email)

	err := service.ForgotPassword(context.Background(), "x@example.com")
	require.Error(test, err)

	var dbUser user.User
	require.NoError(test, database.DB.Where("id = ?", u.ID).First(&dbUser).Error)
	assert.Nil(test, dbUser.PasswordResetTokenHash)
	assert.Nil(test, dbUser.PasswordResetExpiresAt)
}

// En: TestServiceResetPasswordSuccess updates password and clears token.
// Es: TestServiceResetPasswordSuccess actualiza contraseña y limpia token.
func TestServiceResetPasswordSuccess(test *testing.T) {
	stub := &stubPasswordResetNotifier{}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})
	u := seedVerifiedUser(test, "Y", "y@example.com", "oldpassword123")
	seedCredentialsProvider(test, u.ID, u.Email)
	require.NoError(test, service.ForgotPassword(context.Background(), "y@example.com"))

	raw := resetTokenFromLink(stub.lastLink)
	require.NotEmpty(test, raw)
	require.NoError(test, service.ResetPassword(raw, "newpassword456"))

	var dbUser user.User
	require.NoError(test, database.DB.Where("id = ?", u.ID).First(&dbUser).Error)
	assert.Nil(test, dbUser.PasswordResetTokenHash)
	assert.True(test, dbUser.CheckPassword("newpassword456"))
}

// En: TestServiceResetPasswordInvalidToken returns ErrInvalidPasswordResetToken.
// Es: TestServiceResetPasswordInvalidToken devuelve ErrInvalidPasswordResetToken.
func TestServiceResetPasswordInvalidToken(test *testing.T) {
	service := setupServiceTest(test)
	err := service.ResetPassword("not-a-valid-token", "newpassword456")
	assert.ErrorIs(test, err, ErrInvalidPasswordResetToken)
}

// En: TestServiceResetPasswordExpiredToken is rejected.
// Es: TestServiceResetPasswordExpiredToken rechaza token expirado.
func TestServiceResetPasswordExpiredToken(test *testing.T) {
	stub := &stubPasswordResetNotifier{}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})
	u := seedVerifiedUser(test, "Z", "z@example.com", "password123")
	seedCredentialsProvider(test, u.ID, u.Email)
	require.NoError(test, service.ForgotPassword(context.Background(), "z@example.com"))
	raw := resetTokenFromLink(stub.lastLink)
	past := time.Now().Add(-2 * time.Hour)
	require.NoError(test, database.DB.Model(&user.User{}).Where("id = ?", u.ID).Updates(map[string]any{
		"password_reset_expires_at": past,
	}).Error)

	err := service.ResetPassword(raw, "newpassword999")
	assert.ErrorIs(test, err, ErrInvalidPasswordResetToken)
}

// En: TestServiceResetPasswordRevokesRefreshTokens invalidates existing sessions.
// Es: TestServiceResetPasswordRevokesRefreshTokens invalida sesiones existentes.
func TestServiceResetPasswordRevokesRefreshTokens(test *testing.T) {
	stub := &stubPasswordResetNotifier{}
	require.NoError(test, database.InitForTesting())
	require.NoError(test, database.RunMigrations(&user.User{}, &UserAuthProvider{}, &RefreshToken{}, &LoginCredentialLockout{}))
	userRepository := user.NewRepository(database.DB)
	authRepository := NewRepository(database.DB)
	service := NewService(authRepository, userRepository, ServiceOptions{
		JWTSecret:             testJWTSecret,
		VerificationNotifier:  verificationnotify.NoopNotifier{},
		PasswordResetNotifier: stub,
		FrontendURL:           "http://test",
	})
	u := seedVerifiedUser(test, "R", "r@example.com", "password123")
	seedCredentialsProvider(test, u.ID, u.Email)
	pair, err := service.Login("r@example.com", "password123")
	require.NoError(test, err)

	require.NoError(test, service.ForgotPassword(context.Background(), "r@example.com"))
	raw := resetTokenFromLink(stub.lastLink)
	require.NoError(test, service.ResetPassword(raw, "resetpassword1"))

	_, err = service.RefreshTokens(pair.RefreshToken)
	assert.ErrorIs(test, err, ErrInvalidCredentials)
}
