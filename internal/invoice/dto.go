package invoice

// CreateInvoiceRequest is the body for POST /invoices.
type CreateInvoiceRequest struct {
	Number     string `json:"number"      validate:"required,min=1,max=100"`
	Currency   string `json:"currency"    validate:"required,len=3"`
	TotalCents int64  `json:"total_cents" validate:"min=0"`
}
