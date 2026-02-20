package account

import "fmt"

// ErrNotFound is returned when an account is not found.
var ErrNotFound = fmt.Errorf("account not found")

// ErrSlugTaken is returned when the requested slug is already in use.
var ErrSlugTaken = fmt.Errorf("slug already taken")

// ErrMemberNotFound is returned when a membership record is not found.
var ErrMemberNotFound = fmt.Errorf("member not found")

// ErrUserEmailNotVerified is returned when the user's email has not been verified.
var ErrUserEmailNotVerified = fmt.Errorf("user email not verified")
