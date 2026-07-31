package http

import (
	"strconv"

	"erp-system/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo represents error information in the response
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// Meta represents metadata in the response (pagination, etc.)
type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Success sends a successful response
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *fiber.Ctx, message string, data interface{}, meta *Meta) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 Created response
func Created(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Error sends an error response using the app error
func Error(c *fiber.Ctx, err *errors.AppError) error {
	// Get HTTP status from error, default to 500
	status := err.HTTPStatus
	if status == 0 {
		status = fiber.StatusInternalServerError
	}

	errorInfo := &ErrorInfo{
		Code:    string(err.Code),
		Message: err.Message,
		Details: err.Details,
	}

	// Add validation fields if present
	if err.Meta != nil {
		errorInfo.Fields = err.Meta
	}

	return c.Status(status).JSON(Response{
		Success: false,
		Error:   errorInfo,
	})
}

// BadRequest sends a 400 Bad Request error response
func BadRequest(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeBadRequest),
			Message: message,
		},
	})
}

// Unauthorized sends a 401 Unauthorized error response
func Unauthorized(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeUnauthorized),
			Message: message,
		},
	})
}

// Forbidden sends a 403 Forbidden error response
func Forbidden(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeForbidden),
			Message: message,
		},
	})
}

// NotFound sends a 404 Not Found error response
func NotFound(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusNotFound).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeNotFound),
			Message: message,
		},
	})
}

// Conflict sends a 409 Conflict error response
func Conflict(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeConflict),
			Message: message,
		},
	})
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusInternalServerError).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeInternal),
			Message: message,
		},
	})
}

// ValidationError sends a 422 Unprocessable Entity error response with field errors
func ValidationError(c *fiber.Ctx, message string, fields map[string]interface{}) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    string(errors.CodeValidation),
			Message: message,
			Fields:  fields,
		},
	})
}

// ServiceUnavailable sends a 503 Service Unavailable error response
func ServiceUnavailable(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusServiceUnavailable).JSON(Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    "SERVICE_UNAVAILABLE",
			Message: message,
		},
	})
}

// JSON sends a custom JSON response
func JSON(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(data)
}

// ParsePage extracts page/per_page from query params.
// Returns page, perPage, limit (=perPage), offset ((page-1)*perPage).
//
// Defaults are generous: page=1, per_page=500, cap=1000. The high default
// is so legacy callers (E2E specs, the FE's non-paginated list fetches)
// that don't specify `?per_page` still see a realistic tenant's worth of
// rows in one response. Pages that use the shared `<Pagination>` UI
// explicitly pass `per_page=20` (or whatever the row-picker is set to).
func ParsePage(c *fiber.Ctx) (page, perPage, limit, offset int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	perPage, _ = strconv.Atoi(c.Query("per_page", "500"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 500
	}
	if perPage > 1000 {
		perPage = 1000
	}
	return page, perPage, perPage, (page - 1) * perPage
}

// Paginate calculates pagination metadata
func Paginate(page, perPage, total int) *Meta {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
