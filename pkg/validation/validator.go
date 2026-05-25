package validation

import (
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// Get returns the singleton validator with all custom rules registered.
func Get() *validator.Validate {
	once.Do(func() {
		instance = validator.New()
		registerCustomRules(instance)
	})
	return instance
}

// ValidateStruct validates a struct and returns field-level errors or nil.
func ValidateStruct(s interface{}) map[string]string {
	err := Get().Struct(s)
	if err == nil {
		return nil
	}
	fields := make(map[string]string)
	for _, e := range err.(validator.ValidationErrors) {
		fields[strings.ToLower(e.Field())] = fieldMessage(e)
	}
	return fields
}

func fieldMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Too short (minimum " + e.Param() + " characters)"
	case "max":
		return "Too long (maximum " + e.Param() + " characters)"
	case "oneof":
		return "Must be one of: " + e.Param()
	case "slug":
		return "Must contain only lowercase letters, digits, and hyphens"
	case "strongpassword":
		return "Must be at least 8 characters with uppercase, lowercase, digit, and special character"
	case "phone":
		return "Must be a valid E.164 phone number (e.g. +94771234567)"
	case "currency_code":
		return "Must be a valid ISO 4217 currency code (e.g. USD, LKR)"
	case "country_code":
		return "Must be a valid ISO 3166-1 alpha-2 country code (e.g. US, LK)"
	case "alphanum":
		return "Must contain only letters and numbers"
	case "numeric":
		return "Must be a number"
	default:
		return "Invalid value"
	}
}
