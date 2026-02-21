package user

import "fmt"

// En: ErrNotFound is returned when a user is not found.
// Es: ErrNotFound se devuelve cuando no se encuentra un usuario.
var ErrNotFound = fmt.Errorf("user not found")

// En: ErrDuplicateEmail is returned when creating a user with an email that already exists.
// Es: ErrDuplicateEmail se devuelve al crear un usuario con un email que ya existe.
var ErrDuplicateEmail = fmt.Errorf("email already exists")
