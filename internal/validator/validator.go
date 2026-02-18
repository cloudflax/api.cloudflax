package validator

import (
	"github.com/go-playground/validator/v10"
)

// Global validator instance
var validate = validator.New()

// Validate checks a struct for validation tags
func Validate(s interface{}) error {
	return validate.Struct(s)
}
