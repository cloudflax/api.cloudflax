package user

import "fmt"

// ErrNotFound is returned when a user is not found.
var ErrNotFound = fmt.Errorf("user not found")

// ErrDuplicateEmail is returned when creating a user with an email that already exists.
var ErrDuplicateEmail = fmt.Errorf("email already exists")
