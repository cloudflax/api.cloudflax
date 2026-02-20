package invoice

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Repository handles invoice data access, always scoped to an account.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new invoice repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ListInvoice returns all invoices belonging to the given account.
func (r *Repository) ListInvoice(accountID string) ([]Invoice, error) {
	var invoices []Invoice
	if err := r.db.Where("account_id = ?", accountID).Find(&invoices).Error; err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	return invoices, nil
}

// GetInvoice returns an invoice by ID, enforcing that it belongs to the given account.
func (r *Repository) GetInvoice(id, accountID string) (*Invoice, error) {
	var inv Invoice
	if err := r.db.First(&inv, "id = ? AND account_id = ?", id, accountID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get invoice: %w", err)
	}
	return &inv, nil
}

// CreateInvoice persists a new invoice.
func (r *Repository) CreateInvoice(inv *Invoice) error {
	if err := r.db.Create(inv).Error; err != nil {
		return fmt.Errorf("create invoice: %w", err)
	}
	return nil
}
