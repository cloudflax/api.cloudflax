package invoice

import (
	"github.com/gofiber/fiber/v3"
)

// Routes mounts invoice routes on the given router.
// All routes require authentication (authMiddleware) and account membership (accountMiddleware).
func Routes(router fiber.Router, handler *Handler, authMiddleware, accountMiddleware fiber.Handler) {
	invoices := router.Group("/invoices", authMiddleware, accountMiddleware)
	invoices.Get("/", handler.ListInvoice)
	invoices.Get("/:id", handler.GetInvoice)
	invoices.Post("/", handler.CreateInvoice)
}
