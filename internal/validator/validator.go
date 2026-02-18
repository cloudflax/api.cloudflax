package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// FieldError describes a single field-level validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a slice of FieldError returned when struct validation fails.
type ValidationErrors []FieldError

func (e ValidationErrors) Error() string {
	parts := make([]string, len(e))
	for i, fe := range e {
		parts[i] = fe.Field + ": " + fe.Message
	}
	return strings.Join(parts, "; ")
}

// Validate checks a struct against its validation tags.
// On failure it returns ValidationErrors so callers can inspect individual fields.
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fieldErrors := make(ValidationErrors, 0, len(ve))
			for _, fe := range ve {
				fieldErrors = append(fieldErrors, FieldError{
					Field:   fe.Field(),
					Message: fieldMessage(fe),
				})
			}
			return fieldErrors
		}
		return err
	}
	return nil
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at least %s characters", fe.Param())
		}
		return fmt.Sprintf("Must be at least %s", fe.Param())
	case "max":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at most %s characters", fe.Param())
		}
		return fmt.Sprintf("Must be at most %s", fe.Param())
	default:
		return "Is invalid"
	}
}
