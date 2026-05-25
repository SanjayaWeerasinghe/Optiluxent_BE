package validation

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	slugRe  = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	phoneRe = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
)

func registerCustomRules(v *validator.Validate) {
	_ = v.RegisterValidation("slug", validateSlug)
	_ = v.RegisterValidation("strongpassword", validateStrongPassword)
	_ = v.RegisterValidation("phone", validatePhone)
	_ = v.RegisterValidation("currency_code", validateCurrencyCode)
	_ = v.RegisterValidation("country_code", validateCountryCode)
}

// slug: lowercase alphanumeric segments joined by single hyphens
func validateSlug(fl validator.FieldLevel) bool {
	return slugRe.MatchString(fl.Field().String())
}

// strongpassword: min 8 chars, must have upper + lower + digit + special
func validateStrongPassword(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

// phone: E.164 format
func validatePhone(fl validator.FieldLevel) bool {
	return phoneRe.MatchString(fl.Field().String())
}

// currency_code: ISO 4217 (3-letter uppercase)
func validateCurrencyCode(fl validator.FieldLevel) bool {
	code := strings.ToUpper(fl.Field().String())
	_, ok := iso4217[code]
	return ok
}

// country_code: ISO 3166-1 alpha-2 (2-letter uppercase)
func validateCountryCode(fl validator.FieldLevel) bool {
	code := strings.ToUpper(fl.Field().String())
	_, ok := iso3166[code]
	return ok
}

// Subset of ISO 4217 currency codes relevant to this system
var iso4217 = map[string]struct{}{
	"AED": {}, "AUD": {}, "BDT": {}, "BRL": {}, "CAD": {}, "CHF": {}, "CNY": {},
	"EUR": {}, "GBP": {}, "HKD": {}, "IDR": {}, "INR": {}, "JPY": {}, "KRW": {},
	"LKR": {}, "MYR": {}, "MXN": {}, "NPR": {}, "NZD": {}, "PHP": {}, "PKR": {},
	"QAR": {}, "SAR": {}, "SGD": {}, "THB": {}, "TRY": {}, "USD": {}, "VND": {},
	"ZAR": {},
}

// Subset of ISO 3166-1 alpha-2 country codes
var iso3166 = map[string]struct{}{
	"AE": {}, "AU": {}, "BD": {}, "BR": {}, "CA": {}, "CH": {}, "CN": {},
	"DE": {}, "FR": {}, "GB": {}, "HK": {}, "ID": {}, "IN": {}, "JP": {},
	"KR": {}, "LK": {}, "MY": {}, "MX": {}, "NP": {}, "NZ": {}, "PH": {},
	"PK": {}, "QA": {}, "SA": {}, "SG": {}, "TH": {}, "TR": {}, "US": {},
	"VN": {}, "ZA": {},
}
