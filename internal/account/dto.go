package account

// CreateAccountRequest is the request body for POST /accounts.
type CreateAccountRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
	Slug string `json:"slug" validate:"omitempty,min=2,max=100,slug"`
}
