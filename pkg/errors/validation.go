package errors

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	var messages []string
	for _, err := range v.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// FromValidatorError converts validator.ValidationErrors to our ValidationErrors
func FromValidatorError(err error) *AppError {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		var errors []ValidationError

		for _, fieldErr := range validationErrs {
			errors = append(errors, ValidationError{
				Field:   getJSONFieldName(fieldErr.Field()),
				Message: getErrorMessage(fieldErr),
				Tag:     fieldErr.Tag(),
				Value:   fmt.Sprintf("%v", fieldErr.Value()),
			})
		}

		appErr := Validation("Validation failed")
		appErr.Meta = map[string]interface{}{
			"validation_errors": errors,
		}
		return appErr
	}

	return Validation(err.Error())
}

// getJSONFieldName converts struct field name to JSON field name (lowercase first letter)
func getJSONFieldName(field string) string {
	if len(field) == 0 {
		return field
	}
	// Convert first letter to lowercase
	return strings.ToLower(field[:1]) + field[1:]
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(fieldErr validator.FieldError) string {
	field := getJSONFieldName(fieldErr.Field())

	switch fieldErr.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fieldErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fieldErr.Param())
	case "len":
		return fmt.Sprintf("%s must be %s characters long", field, fieldErr.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fieldErr.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fieldErr.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fieldErr.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fieldErr.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fieldErr.Param())
	case "eqfield":
		return fmt.Sprintf("%s must equal %s", field, fieldErr.Param())
	case "nefield":
		return fmt.Sprintf("%s must not equal %s", field, fieldErr.Param())
	case "unique":
		return fmt.Sprintf("%s must contain unique values", field)
	default:
		return fmt.Sprintf("%s failed validation for tag: %s", field, fieldErr.Tag())
	}
}

// NewValidationError creates a single validation error
func NewValidationError(field, message string) *AppError {
	appErr := Validation("Validation failed")
	appErr.Meta = map[string]interface{}{
		"validation_errors": []ValidationError{
			{
				Field:   field,
				Message: message,
			},
		},
	}
	return appErr
}

// AddValidationError adds a validation error to an existing AppError
func AddValidationError(appErr *AppError, field, message string) *AppError {
	if appErr.Meta == nil {
		appErr.Meta = make(map[string]interface{})
	}

	errors, ok := appErr.Meta["validation_errors"].([]ValidationError)
	if !ok {
		errors = []ValidationError{}
	}

	errors = append(errors, ValidationError{
		Field:   field,
		Message: message,
	})

	appErr.Meta["validation_errors"] = errors
	return appErr
}
