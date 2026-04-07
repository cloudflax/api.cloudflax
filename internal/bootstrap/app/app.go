package app

import (
	"github.com/cloudflax/api.cloudflax/internal/bootstrap/config"
	"github.com/cloudflax/api.cloudflax/internal/bootstrap/server"
	"github.com/cloudflax/api.cloudflax/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

// fiberAppConfig builds Fiber settings (proxy trust for correct client IP behind load balancers).
func fiberAppConfig(cfg *config.Config) fiber.Config {
	c := fiber.Config{
		TrustProxy:      cfg.TrustProxy,
		ProxyHeader:     cfg.ProxyHeader,
		TrustProxyConfig: fiber.TrustProxyConfig{
			Private:  cfg.TrustProxyPrivate,
			Loopback: cfg.TrustProxyLoopback,
		},
	}
	return c
}

// Run starts the Fiber server with the loaded configuration.
func Run(cfg *config.Config) error {
	app := fiber.New(fiberAppConfig(cfg))

	app.Use(middleware.Logger())
	app.Use(middleware.CORS(cfg.FrontendURL))
	server.Mount(app, cfg)

	return app.Listen(":" + cfg.Port)
}
