package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// General errors
	CodeInternal        ErrorCode = "INTERNAL_ERROR"
	CodeBadRequest      ErrorCode = "BAD_REQUEST"
	CodeNotFound        ErrorCode = "NOT_FOUND"
	CodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	CodeForbidden       ErrorCode = "FORBIDDEN"
	CodeConflict        ErrorCode = "CONFLICT"
	CodeValidation      ErrorCode = "VALIDATION_ERROR"
	CodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Authentication errors
	CodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	CodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	CodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	CodeTokenMissing       ErrorCode = "TOKEN_MISSING"

	// Database errors
	CodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	CodeRecordNotFound     ErrorCode = "RECORD_NOT_FOUND"
	CodeDuplicateEntry     ErrorCode = "DUPLICATE_ENTRY"
	CodeForeignKeyViolation ErrorCode = "FOREIGN_KEY_VIOLATION"

	// Business logic errors
	CodeInsufficientPermission ErrorCode = "INSUFFICIENT_PERMISSION"
	CodeResourceNotFound       ErrorCode = "RESOURCE_NOT_FOUND"
	CodeResourceAlreadyExists  ErrorCode = "RESOURCE_ALREADY_EXISTS"
	CodeInvalidOperation       ErrorCode = "INVALID_OPERATION"
	CodeInvalidState           ErrorCode = "INVALID_STATE"
)

// AppError represents an application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Err        error                  `json:"-"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithMeta adds metadata to the error
func (e *AppError) WithMeta(key string, value interface{}) *AppError {
	if e.Meta == nil {
		e.Meta = make(map[string]interface{})
	}
	e.Meta[key] = value
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// NewWithStatus creates a new AppError with custom HTTP status
func NewWithStatus(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getDefaultHTTPStatus(code),
		Err:        err,
	}
}

// getDefaultHTTPStatus returns the default HTTP status for an error code
func getDefaultHTTPStatus(code ErrorCode) int {
	statusMap := map[ErrorCode]int{
		CodeInternal:               http.StatusInternalServerError,
		CodeBadRequest:             http.StatusBadRequest,
		CodeNotFound:               http.StatusNotFound,
		CodeUnauthorized:           http.StatusUnauthorized,
		CodeForbidden:              http.StatusForbidden,
		CodeConflict:               http.StatusConflict,
		CodeValidation:             http.StatusBadRequest,
		CodeTooManyRequests:        http.StatusTooManyRequests,
		CodeInvalidCredentials:     http.StatusUnauthorized,
		CodeTokenExpired:           http.StatusUnauthorized,
		CodeTokenInvalid:           http.StatusUnauthorized,
		CodeTokenMissing:           http.StatusUnauthorized,
		CodeDatabaseError:          http.StatusInternalServerError,
		CodeRecordNotFound:         http.StatusNotFound,
		CodeDuplicateEntry:         http.StatusConflict,
		CodeForeignKeyViolation:    http.StatusBadRequest,
		CodeInsufficientPermission: http.StatusForbidden,
		CodeResourceNotFound:       http.StatusNotFound,
		CodeResourceAlreadyExists:  http.StatusConflict,
		CodeInvalidOperation:       http.StatusBadRequest,
		CodeInvalidState:           http.StatusBadRequest,
	}

	if status, ok := statusMap[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// Predefined error constructors

// Internal creates an internal server error
func Internal(message string) *AppError {
	return New(CodeInternal, message)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message)
}

// NotFound creates a not found error
func NotFound(message string) *AppError {
	return New(CodeNotFound, message)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message)
}

// Conflict creates a conflict error
func Conflict(message string) *AppError {
	return New(CodeConflict, message)
}

// Validation creates a validation error
func Validation(message string) *AppError {
	return New(CodeValidation, message)
}

// InvalidCredentials creates an invalid credentials error
func InvalidCredentials() *AppError {
	return New(CodeInvalidCredentials, "Invalid email or password")
}

// TokenExpired creates a token expired error
func TokenExpired() *AppError {
	return New(CodeTokenExpired, "Token has expired")
}

// TokenInvalid creates an invalid token error
func TokenInvalid() *AppError {
	return New(CodeTokenInvalid, "Invalid token")
}

// TokenMissing creates a missing token error
func TokenMissing() *AppError {
	return New(CodeTokenMissing, "Authorization token is missing")
}

// DatabaseError creates a database error
func DatabaseError(err error) *AppError {
	return Wrap(err, CodeDatabaseError, "Database operation failed")
}

// RecordNotFound creates a record not found error
func RecordNotFound(resource string) *AppError {
	return New(CodeRecordNotFound, fmt.Sprintf("%s not found", resource))
}

// DuplicateEntry creates a duplicate entry error
func DuplicateEntry(field string) *AppError {
	return New(CodeDuplicateEntry, fmt.Sprintf("%s already exists", field))
}

// InsufficientPermission creates an insufficient permission error
func InsufficientPermission(action string) *AppError {
	return New(CodeInsufficientPermission, fmt.Sprintf("Insufficient permission to %s", action))
}

// InvalidOperation creates an invalid operation error
func InvalidOperation(message string) *AppError {
	return New(CodeInvalidOperation, message)
}

// Is checks if an error matches a specific error code
func Is(err error, code ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// GetCode extracts the error code from an error
func GetCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return CodeInternal
}

// GetHTTPStatus extracts the HTTP status from an error
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
