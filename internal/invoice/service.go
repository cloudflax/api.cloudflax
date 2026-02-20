package invoice

// Service handles invoice business logic.
type Service struct {
	repository *Repository
}

// NewService creates a new invoice service.
func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// ListInvoice returns all invoices for the given account.
func (s *Service) ListInvoice(accountID string) ([]Invoice, error) {
	return s.repository.ListInvoice(accountID)
}

// GetInvoice returns a single invoice by ID, scoped to the given account.
func (s *Service) GetInvoice(id, accountID string) (*Invoice, error) {
	return s.repository.GetInvoice(id, accountID)
}

// CreateInvoice creates a new invoice within the given account.
func (s *Service) CreateInvoice(accountID, number, currency string, totalCents int64) (*Invoice, error) {
	inv := &Invoice{
		AccountID:  accountID,
		Number:     number,
		Status:     StatusDraft,
		TotalCents: totalCents,
		Currency:   currency,
	}
	if err := s.repository.CreateInvoice(inv); err != nil {
		return nil, err
	}
	return inv, nil
}
