package invoice

import "errors"

// ErrNotFound is returned when an invoice does not exist or does not belong to the account.
var ErrNotFound = errors.New("invoice not found")
