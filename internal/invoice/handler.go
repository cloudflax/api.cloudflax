package invoice

import (
	"errors"
	"log/slog"

	runtimeError "github.com/cloudflax/api.cloudflax/internal/shared/runtimeerror"
	"github.com/cloudflax/api.cloudflax/internal/shared/requestctx"
	"github.com/cloudflax/api.cloudflax/internal/validator"
	"github.com/gofiber/fiber/v3"
)

// Handler handles HTTP requests for invoices.
type Handler struct {
	service *Service
}

// NewHandler creates a new invoice handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListInvoice handles GET /invoices.
// Returns all invoices scoped to the account in the request context.
func (h *Handler) ListInvoice(c fiber.Ctx) error {
	rctx, err := requestctx.FromFiber(c)
	if err != nil {
		return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeUnauthorized, "Unauthorized")
	}

	invoices, err := h.service.ListInvoice(rctx.AccountID)
	if err != nil {
		slog.Error("list invoices", "account_id", rctx.AccountID, "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Failed to list invoices")
	}

	return c.JSON(fiber.Map{"data": invoices})
}

// GetInvoice handles GET /invoices/:id.
// Returns a single invoice scoped to the account in the request context.
func (h *Handler) GetInvoice(c fiber.Ctx) error {
	rctx, err := requestctx.FromFiber(c)
	if err != nil {
		return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeUnauthorized, "Unauthorized")
	}

	id := c.Params("id")
	inv, err := h.service.GetInvoice(id, rctx.AccountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return runtimeError.Respond(c, fiber.StatusNotFound, runtimeError.CodeInvoiceNotFound, "Invoice not found")
		}
		slog.Error("get invoice", "id", id, "account_id", rctx.AccountID, "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Failed to get invoice")
	}

	return c.JSON(fiber.Map{"data": inv})
}

// CreateInvoice handles POST /invoices.
// Creates a new invoice scoped to the account in the request context.
func (h *Handler) CreateInvoice(c fiber.Ctx) error {
	rctx, err := requestctx.FromFiber(c)
	if err != nil {
		return runtimeError.Respond(c, fiber.StatusUnauthorized, runtimeError.CodeUnauthorized, "Unauthorized")
	}

	var req CreateInvoiceRequest
	if err := c.Bind().Body(&req); err != nil {
		slog.Debug("create invoice bind error", "error", err)
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeInvalidRequestBody, "Invalid request body")
	}

	if err := validator.Validate(req); err != nil {
		slog.Debug("create invoice validation error", "error", err)
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return runtimeError.RespondWithDetails(
				c, fiber.StatusUnprocessableEntity, runtimeError.CodeValidationError,
				"Validation failed", toErrorDetails(ve),
			)
		}
		return runtimeError.Respond(c, fiber.StatusBadRequest, runtimeError.CodeValidationError, err.Error())
	}

	inv, err := h.service.CreateInvoice(rctx.AccountID, req.Number, req.Currency, req.TotalCents)
	if err != nil {
		slog.Error("create invoice", "account_id", rctx.AccountID, "error", err)
		return runtimeError.Respond(c, fiber.StatusInternalServerError, runtimeError.CodeInternalServerError, "Failed to create invoice")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": inv})
}

// toErrorDetails converts validator.ValidationErrors to runtimeError.ErrorDetail slice.
func toErrorDetails(ve validator.ValidationErrors) []runtimeError.ErrorDetail {
	details := make([]runtimeError.ErrorDetail, len(ve))
	for i, fe := range ve {
		details[i] = runtimeError.ErrorDetail{
			Field:   fe.Field,
			Message: fe.Message,
		}
	}
	return details
}
