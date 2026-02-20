package server

import (
	"github.com/cloudflax/api.cloudflax/internal/account"
	"github.com/cloudflax/api.cloudflax/internal/auth"
	"github.com/cloudflax/api.cloudflax/internal/config"
	"github.com/cloudflax/api.cloudflax/internal/invoice"
	"github.com/cloudflax/api.cloudflax/internal/shared/database"
	"github.com/cloudflax/api.cloudflax/internal/shared/middleware"
	"github.com/cloudflax/api.cloudflax/internal/user"
	"github.com/gofiber/fiber/v3"
)

// Mount mounts all routes on the Fiber app.
func Mount(app *fiber.App, cfg *config.Config) {
	app.Get("/", Home)
	app.Get("/health", Health())

	authRepository := auth.NewRepository(database.DB)
	userRepository := user.NewRepository(database.DB)
	authService := auth.NewService(authRepository, userRepository, cfg.JWTSecret)
	authHandler := auth.NewHandler(authService)

	requireAuth := middleware.RequireAuth(authService)

	auth.Routes(app, authHandler, requireAuth)

	userService := user.NewService(userRepository).WithTokenRevoker(authRepository)
	userHandler := user.NewHandler(userService)
	user.Routes(app, userHandler, requireAuth)

	accountRepository := account.NewRepository(database.DB)
	accountService := account.NewService(accountRepository, userRepository)
	accountHandler := account.NewHandler(accountService)
	account.Routes(app, accountHandler, requireAuth)

	requireAccountMember := middleware.RequireAccountMember(accountRepository)

	invoiceRepository := invoice.NewRepository(database.DB)
	invoiceService := invoice.NewService(invoiceRepository)
	invoiceHandler := invoice.NewHandler(invoiceService)
	invoice.Routes(app, invoiceHandler, requireAuth, requireAccountMember)
}
