package server

import (
	"context"
	"log/slog"

	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/auth"
	"github.com/cloudflax/api.cloudflax/internal/bootstrap/config"
	"github.com/cloudflax/api.cloudflax/internal/invoice"
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/shared/email"
	"github.com/cloudflax/api.cloudflax/internal/shared/middleware"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
)

// Mount mounts all routes on the Fiber app.
func Mount(app *fiber.App, cfg *config.Config) {
	app.Get("/", Home)
	app.Get("/health", Health())

	emailSender := newEmailSender(cfg)

	authRepository := auth.NewRepository(database.DB)
	userRepository := user.NewRepository(database.DB)
	accountRepository := account.NewRepository(database.DB)
	accountService := account.NewService(accountRepository, userRepository)

	authService := auth.NewService(authRepository, userRepository, cfg.JWTSecret, emailSender, cfg.AppURL)
	authHandler := auth.NewHandler(authService)
	requireAuth := middleware.RequireAuth(authService)
	auth.Routes(app, authHandler, requireAuth)

	userService := user.NewService(userRepository).WithTokenRevoker(authRepository)
	userHandler := user.NewHandler(userService).WithAccountLister(&accountListerAdapter{service: accountService})
	user.Routes(app, userHandler, requireAuth)

	accountHandler := account.NewHandler(accountService)
	requireAccountMember := middleware.RequireAccountMember(accountRepository)
	account.Routes(app, accountHandler, requireAuth)

	invoiceRepository := invoice.NewRepository(database.DB)
	invoiceService := invoice.NewService(invoiceRepository)
	invoiceHandler := invoice.NewHandler(invoiceService)
	invoice.Routes(app, invoiceHandler, requireAuth, requireAccountMember)
}

// accountListerAdapter adapts the account.Service to the user.AccountLister interface.
type accountListerAdapter struct {
	service *account.Service
}

func (a *accountListerAdapter) ListAccountsForUser(userID string) ([]user.AccountSummary, error) {
	accounts, err := a.service.ListAccountsForUser(userID)
	if err != nil {
		return nil, err
	}

	result := make([]user.AccountSummary, len(accounts))
	for i, acc := range accounts {
		result[i] = user.AccountSummary{
			ID:   acc.ID,
			Name: acc.Name,
			Slug: acc.Slug,
		}
	}
	return result, nil
}

// newEmailSender builds an SES-backed TemplatedSender from the loaded config.
// Falls back to a no-op sender and logs a warning if SES cannot be initialised.
func newEmailSender(cfg *config.Config) email.TemplatedSender {
	sender, err := email.NewSESSender(context.Background(), email.SESSenderOptions{
		EndpointURL:     cfg.AWSEndpointURL,
		Region:          cfg.AWSRegion,
		AccessKeyID:     cfg.AWSAccessKeyID,
		SecretAccessKey: cfg.AWSSecretAccessKey,
		FromAddress:     cfg.SESFromAddress,
	})
	if err != nil {
		slog.Warn("failed to initialise SES sender, falling back to noop", "error", err)
		return &email.NoopSender{}
	}
	return sender
}
