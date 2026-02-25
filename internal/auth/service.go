package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/cloudflax/api.cloudflax/internal/shared/email"
	"github.com/cloudflax/api.cloudflax/internal/user"
)

// En: accessTokenDuration is the duration of the access token.
// Es: accessTokenDuration es el tiempo de duración del token de acceso.
const (
	accessTokenDuration       = 15 * time.Minute
	refreshTokenDuration      = 7 * 24 * time.Hour
	refreshTokenBytes         = 32
	verificationTokenDuration = 24 * time.Hour
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
	GetUserByEmail(email string) (*user.User, error)
	Create(u *user.User) error
	Update(u *user.User) error
}

// En: Service handles the business logic of authentication.
// Es: Service maneja la lógica de negocios de la autenticación.
type Service struct {
	repository     *Repository
	userRepository UserRepository
	jwtSecret      []byte
	emailSender    email.Sender
}

// En: NewService creates a new authentication service.
// Es: NewService crea un nuevo servicio de autenticación.
func NewService(repository *Repository, userRepository UserRepository, jwtSecret string, emailSender email.Sender) *Service {
	return &Service{
		repository:     repository,
		userRepository: userRepository,
		jwtSecret:      []byte(jwtSecret),
		emailSender:    emailSender,
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

	if err := service.emailSender.SendVerificationEmail(u.Email, u.Name, token); err != nil {
		slog.Error("send verification email after register", "email", u.Email, "error", err)
	}

	return u, token, nil
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

	if err := service.emailSender.SendVerificationEmail(u.Email, u.Name, token); err != nil {
		slog.Error("send verification email after resend", "email", u.Email, "error", err)
	}

	return token, nil
}

// En: Login verifies the credentials and emits a token pair in case of success.
// Es: Login verifica las credenciales y emite un par de tokens en caso de éxito.
func (service *Service) Login(email, password string) (*TokenPair, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	u, err := service.userRepository.GetUserByEmail(normalizedEmail)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !u.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}
	if !u.IsEmailVerified() {
		return nil, ErrEmailNotVerified
	}
	return service.generateTokenPair(u)
}

// En: RefreshTokens validates an existing refresh token, revokes it (rotation) and emits a new token pair.
// Es: RefreshTokens valida un token de actualización existente, lo revoca (rotación) y emite un nuevo par de tokens.
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
	if !u.IsEmailVerified() {
		return nil, ErrEmailNotVerified
	}
	return service.generateTokenPair(u)
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
