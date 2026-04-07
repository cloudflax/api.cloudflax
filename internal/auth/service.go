package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/cloudflax/api.cloudflax/internal/shared/verificationnotify"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"gorm.io/gorm"
)

const (
	defaultAccessTokenDuration = 15 * time.Minute
	refreshTokenDuration       = 7 * 24 * time.Hour
	refreshTokenBytes          = 32
	verificationTokenDuration  = 24 * time.Hour
	forgotPasswordTokenDuration = time.Hour
)

// En: Claims holds the JWT payload for access tokens.
// Es: Claims contiene el payload del JWT para los tokens de acceso.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// En: TokenPair holds the access token and refresh token issued after a successful login or refresh.
// Es: TokenPair contiene el token de acceso y el token de actualización emitidos después de un inicio de sesión o actualización exitosos.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// En: UserRepository is the subset of the user repository that the authentication service depends on.
// Es: UserRepository es el subconjunto del repositorio de usuarios en el que depende el servicio de autenticación.
type UserRepository interface {
	GetUser(id string) (*user.User, error)
	GetUserTx(tx *gorm.DB, id string) (*user.User, error)
	GetUserByEmail(email string) (*user.User, error)
	Create(u *user.User) error
	Update(u *user.User) error
}

// En: ServiceOptions configures JWT signing, verification and password-reset email delivery, and frontend URL for auth links.
// Es: ServiceOptions configura la firma JWT, el envío de correos de verificación y de recuperación de contraseña, y la URL del front para enlaces de auth.
type ServiceOptions struct {
	JWTSecret            string
	VerificationNotifier verificationnotify.Notifier
	// PasswordResetNotifier sends forgot-password emails (async Lambda → SES). Nil defaults to noop.
	PasswordResetNotifier verificationnotify.PasswordResetEmailNotifier
	FrontendURL          string
	// AccessTokenDuration is the JWT access token lifetime; zero defaults to 15 minutes.
	AccessTokenDuration time.Duration
}

// En: Service handles the business logic of authentication.
// Es: Service maneja la lógica de negocios de la autenticación.
type Service struct {
	repository                *Repository
	userRepository            UserRepository
	jwtSecret                 []byte
	verificationNotifier      verificationnotify.Notifier
	passwordResetNotifier     verificationnotify.PasswordResetEmailNotifier
	frontendURL               string
	accessTokenDuration       time.Duration
}

// En: NewService creates a new authentication service.
// Es: NewService crea un nuevo servicio de autenticación.
func NewService(repository *Repository, userRepository UserRepository, opts ServiceOptions) *Service {
	notifier := opts.VerificationNotifier
	if notifier == nil {
		notifier = verificationnotify.NoopNotifier{}
	}
	resetNotifier := opts.PasswordResetNotifier
	if resetNotifier == nil {
		resetNotifier = verificationnotify.NoopPasswordResetEmailNotifier{}
	}
	accessDur := opts.AccessTokenDuration
	if accessDur <= 0 {
		accessDur = defaultAccessTokenDuration
	}
	return &Service{
		repository:            repository,
		userRepository:        userRepository,
		jwtSecret:             []byte(opts.JWTSecret),
		verificationNotifier:  notifier,
		passwordResetNotifier: resetNotifier,
		frontendURL:           strings.TrimSuffix(strings.TrimSpace(opts.FrontendURL), "/"),
		accessTokenDuration:   accessDur,
	}
}

// En: Register creates a new user with an email/password credential and a pending email verification token.
// Es: Register crea un nuevo usuario con una credencial de correo electrónico/contraseña y un token de verificación de correo electrónico pendiente.
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

	if err := service.sendVerificationEmail(context.Background(), u.Email, u.Name, token); err != nil {
		slog.Error("send verification email after register", "email", u.Email, "error", err)
	}

	return u, token, nil
}

// sendVerificationEmail enqueues verification delivery (async Lambda) with name and verification link.
func (service *Service) sendVerificationEmail(ctx context.Context, toAddress, toName, token string) error {
	if service.frontendURL == "" {
		return fmt.Errorf("frontend URL is required to build verification link")
	}
	link := fmt.Sprintf("%s/auth/verify-email?token=%s", service.frontendURL, token)
	return service.verificationNotifier.NotifyVerificationEmail(ctx, toAddress, toName, link)
}

// En: VerifyEmail marks the user's email as verified using the token previously issued during registration (or after a verification resend request).
// Es: VerifyEmail marca el correo electrónico del usuario como verificado usando el token previamente emitido durante el registro (o después de una solicitud de reenvío de verificación).
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

// En: ResendVerification generates a new email verification token for the given email.
// Es: ResendVerification genera un nuevo token de verificación de correo electrónico para el correo electrónico dado.
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

	if err := service.sendVerificationEmail(context.Background(), u.Email, u.Name, token); err != nil {
		slog.Error("send verification email after resend", "email", u.Email, "error", err)
		return "", fmt.Errorf("send verification email after resend: %w", err)
	}

	return token, nil
}

// En: PeekEmailVerificationToken returns the current verification token without sending email or rotating it.
// Es: PeekEmailVerificationToken devuelve el token de verificación actual sin enviar correo ni rotarlo.
func (service *Service) PeekEmailVerificationToken(email string) (string, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	u, err := service.userRepository.GetUserByEmail(normalizedEmail)
	if err != nil {
		return "", err
	}
	if u.IsEmailVerified() {
		return "", ErrEmailAlreadyVerified
	}
	if u.EmailVerificationToken == nil || strings.TrimSpace(*u.EmailVerificationToken) == "" {
		return "", ErrNoPendingVerificationToken
	}
	if u.EmailVerificationExpiresAt != nil && time.Now().After(*u.EmailVerificationExpiresAt) {
		return "", ErrInvalidVerificationToken
	}
	return *u.EmailVerificationToken, nil
}

// En: ForgotPassword issues a one-time password reset token and enqueues the reset email for eligible users.
// Returns nil without sending when the user does not exist, is unverified, or has no credentials provider (enumeration-safe).
// Es: ForgotPassword emite un token de un solo uso y encola el correo de recuperación para usuarios elegibles.
// Devuelve nil sin enviar si el usuario no existe, no está verificado o no tiene proveedor credentials (sin filtrar enumeración).
func (service *Service) ForgotPassword(ctx context.Context, email string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		slog.Debug("forgot password skipped: empty email after normalization")
		return nil
	}

	u, err := service.userRepository.GetUserByEmail(normalizedEmail)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			slog.Info("forgot password skipped: user not found", "email", normalizedEmail)
			return nil
		}
		return fmt.Errorf("get user for forgot password: %w", err)
	}
	if !u.IsEmailVerified() {
		slog.Info("forgot password skipped: user email not verified", "user_id", u.ID, "email", u.Email)
		return nil
	}

	if _, err := service.repository.FindByProviderAndSubject(ProviderCredentials, normalizedEmail); err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			slog.Info("forgot password skipped: credentials provider not found", "user_id", u.ID, "email", u.Email)
			return nil
		}
		return fmt.Errorf("find credentials provider: %w", err)
	}

	if service.frontendURL == "" {
		return fmt.Errorf("frontend URL is required to build password reset link")
	}

	rawToken, err := generateSecureToken()
	if err != nil {
		return fmt.Errorf("generate password reset token: %w", err)
	}
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(forgotPasswordTokenDuration)
	u.PasswordResetTokenHash = &tokenHash
	u.PasswordResetExpiresAt = &expiresAt
	if err := service.userRepository.Update(u); err != nil {
		return fmt.Errorf("store password reset token: %w", err)
	}

	link := fmt.Sprintf("%s/auth/reset-password?token=%s", service.frontendURL, rawToken)
	expiresIn := formatPasswordResetExpiresIn(forgotPasswordTokenDuration)
	if err := service.passwordResetNotifier.NotifyPasswordResetEmail(ctx, u.Email, u.Name, link, expiresIn); err != nil {
		u.PasswordResetTokenHash = nil
		u.PasswordResetExpiresAt = nil
		if rollbackErr := service.userRepository.Update(u); rollbackErr != nil {
			slog.Error("rollback password reset token after notify failure", "user_id", u.ID, "error", rollbackErr)
		}
		slog.Error("send forgot-password email", "email", u.Email, "error", err)
		return fmt.Errorf("send forgot-password email: %w", err)
	}
	slog.Info("forgot password email queued", "user_id", u.ID, "email", u.Email)
	return nil
}

// En: ResetPassword sets a new password using a valid one-time reset token and revokes all refresh tokens.
// Es: ResetPassword establece una nueva contraseña con un token de recuperación válido de un solo uso y revoca todos los refresh tokens.
func (service *Service) ResetPassword(rawToken, newPassword string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return ErrInvalidPasswordResetToken
	}
	tokenHash := hashToken(rawToken)
	u, err := service.repository.FindByPasswordResetTokenHash(tokenHash)
	if err != nil {
		return err
	}
	if u.PasswordResetExpiresAt == nil || time.Now().After(*u.PasswordResetExpiresAt) {
		return ErrInvalidPasswordResetToken
	}
	if err := u.SetPassword(newPassword); err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	u.PasswordResetTokenHash = nil
	u.PasswordResetExpiresAt = nil
	if err := service.userRepository.Update(u); err != nil {
		return fmt.Errorf("update user password after reset: %w", err)
	}
	if err := service.repository.RevokeAllByUserID(u.ID); err != nil {
		return fmt.Errorf("revoke refresh tokens after password reset: %w", err)
	}
	if err := service.repository.ClearLoginCredentialLockout(strings.ToLower(strings.TrimSpace(u.Email))); err != nil {
		return fmt.Errorf("clear login lockout after password reset: %w", err)
	}
	return nil
}

func formatPasswordResetExpiresIn(d time.Duration) string {
	minutes := int(d.Round(time.Minute) / time.Minute)
	if minutes < 1 {
		minutes = 1
	}
	return fmt.Sprintf("%d minutes", minutes)
}

// En: Login verifies the credentials and emits a token pair in case of success.
// Es: Login verifica las credenciales y emite un par de tokens en caso de éxito.
func (service *Service) Login(email, password string) (*TokenPair, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	lockRemain, err := service.repository.LoginCredentialLockRetryAfter(normalizedEmail)
	if err != nil {
		return nil, fmt.Errorf("login credential lockout: %w", err)
	}
	if lockRemain > 0 {
		return nil, &CredentialsLockedError{RetryAfter: lockRemain}
	}

	u, err := service.userRepository.GetUserByEmail(normalizedEmail)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			if recErr := service.repository.RecordFailedLoginCredentialAttempt(normalizedEmail); recErr != nil {
				return nil, fmt.Errorf("record failed login: %w", recErr)
			}
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if !u.CheckPassword(password) {
		if recErr := service.repository.RecordFailedLoginCredentialAttempt(normalizedEmail); recErr != nil {
			return nil, fmt.Errorf("record failed login: %w", recErr)
		}
		return nil, ErrInvalidCredentials
	}
	if !u.IsEmailVerified() {
		return nil, ErrEmailNotVerified
	}
	if err := service.repository.ClearLoginCredentialLockout(normalizedEmail); err != nil {
		return nil, err
	}
	return service.generateTokenPair(u)
}

// En: RefreshTokens validates an existing refresh token, revokes it (rotation) and emits a new token pair.
// Es: RefreshTokens valida un token de actualización existente, lo revoca (rotación) y emite un nuevo par de tokens.
func (service *Service) RefreshTokens(rawRefreshToken string) (*TokenPair, error) {
	rawRefreshToken = strings.TrimSpace(rawRefreshToken)
	if looksLikeJWT(rawRefreshToken) {
		return nil, ErrJWTUsedAsRefreshToken
	}

	tokenHash := hashToken(rawRefreshToken)

	var pair *TokenPair
	err := service.repository.Transaction(func(tx *gorm.DB) error {
		stored, err := service.repository.ConsumeRefreshByTokenHash(tx, tokenHash)
		if err != nil {
			if errors.Is(err, ErrTokenNotFound) {
				return ErrInvalidCredentials
			}
			return err
		}

		u, err := service.userRepository.GetUserTx(tx, stored.UserID)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				return ErrInvalidCredentials
			}
			return err
		}
		if !u.IsEmailVerified() {
			return ErrEmailNotVerified
		}

		pair, err = service.persistTokenPair(tx, u)
		return err
	})
	if err != nil {
		return nil, err
	}
	return pair, nil
}

// En: Logout revokes all active refresh tokens for the given user.
// Es: Logout revoca todos los tokens de actualización activos para el usuario dado.
func (service *Service) Logout(userID string) error {
	return service.repository.RevokeAllByUserID(userID)
}

// En: ValidateAccessToken validates and analyzes a JWT access token.
// Es: ValidateAccessToken analiza y valida un token de acceso JWT.
func (service *Service) ValidateAccessToken(tokenString string) (string, string, error) {
	claims, err := service.parseAccessToken(tokenString)
	if err != nil {
		return "", "", err
	}
	return claims.UserID, claims.Email, nil
}

// En: parseAccessToken analyzes the JWT and returns the complete Claims struct.
// Es: parseAccessToken analiza el JWT y devuelve el struct Claims completo.
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

// En: generateTokenPair creates and stores a new access token and refresh token pair for the given user.
// Es: generateTokenPair crea y almacena un nuevo par de tokens de acceso y actualización para el usuario dado.
func (service *Service) generateTokenPair(u *user.User) (*TokenPair, error) {
	return service.persistTokenPair(service.repository.db, u)
}

// En: persistTokenPair signs access JWT and inserts refresh metadata using db (or an active transaction).
// Es: persistTokenPair firma el access JWT e inserta el refresh usando db o una transacción activa.
func (service *Service) persistTokenPair(db *gorm.DB, u *user.User) (*TokenPair, error) {
	expiresAt := time.Now().Add(service.accessTokenDuration)
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
	if err := db.Create(refreshRecord).Error; err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
	}, nil
}

// En: signAccessToken builds and signs a JWT for the given user.
// Es: signAccessToken construye y firma un JWT para el usuario dado.
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

// En: generateSecureToken creates a cryptographically random hex-encoded token.
// Es: generateSecureToken crea un token hex-codificado de forma criptográficamente aleatoria.
func generateSecureToken() (string, error) {
	bytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// En: hashToken returns the SHA-256 hex hash of a raw token string.
// Es: hashToken devuelve el hash SHA-256 hex de una cadena de token sin procesar.
func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// looksLikeJWT reports whether s has the typical three Base64URL segments of a JWT.
// Refresh tokens in this API are opaque hex strings without dots.
func looksLikeJWT(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
	}
	return true
}
