package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page    int
	PerPage int
	Offset  int
}

// GetPagination extracts pagination parameters from request
func GetPagination(c *fiber.Ctx) *PaginationParams {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))

	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	return &PaginationParams{
		Page:    page,
		PerPage: perPage,
		Offset:  offset,
	}
}

// GetIntParam extracts an integer parameter from the URL
func GetIntParam(c *fiber.Ctx, key string) (int, error) {
	param := c.Params(key)
	return strconv.Atoi(param)
}

// GetStringParam extracts a string parameter from the URL
func GetStringParam(c *fiber.Ctx, key string) string {
	return c.Params(key)
}

// GetQuery extracts a query parameter with a default value
func GetQuery(c *fiber.Ctx, key, defaultValue string) string {
	return c.Query(key, defaultValue)
}

// GetIntQuery extracts an integer query parameter with a default value
func GetIntQuery(c *fiber.Ctx, key string, defaultValue int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return defaultValue
	}
	return value
}

// GetBoolQuery extracts a boolean query parameter with a default value
func GetBoolQuery(c *fiber.Ctx, key string, defaultValue bool) bool {
	value, err := strconv.ParseBool(c.Query(key))
	if err != nil {
		return defaultValue
	}
	return value
}

// BindJSON binds JSON body to a struct
func BindJSON(c *fiber.Ctx, obj interface{}) error {
	return c.BodyParser(obj)
}

// GetHeader extracts a header value
func GetHeader(c *fiber.Ctx, key string) string {
	return c.Get(key)
}

// GetAuthToken extracts the authorization token from the header
func GetAuthToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

// GetUserID extracts the user ID from context (set by auth middleware)
func GetUserID(c *fiber.Ctx) uint {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return 0
	}
	return userID
}

// SetUserID sets the user ID in context
func SetUserID(c *fiber.Ctx, userID uint) {
	c.Locals("user_id", userID)
}

// GetUserRole extracts the user role from context (set by auth middleware)
func GetUserRole(c *fiber.Ctx) string {
	role, ok := c.Locals("user_role").(string)
	if !ok {
		return ""
	}
	return role
}

// SetUserRole sets the user role in context
func SetUserRole(c *fiber.Ctx, role string) {
	c.Locals("user_role", role)
}

// GetRequestID extracts the request ID from context
func GetRequestID(c *fiber.Ctx) string {
	requestID, ok := c.Locals("request_id").(string)
	if !ok {
		return ""
	}
	return requestID
}

// SetRequestID sets the request ID in context
func SetRequestID(c *fiber.Ctx, requestID string) {
	c.Locals("request_id", requestID)
}

// GetClientIP gets the client IP address
func GetClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header first
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	// Check X-Real-IP header
	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to remote IP
	return c.IP()
}

// GetUserAgent gets the user agent string
func GetUserAgent(c *fiber.Ctx) string {
	return c.Get("User-Agent")
}

// IsAjaxRequest checks if the request is an AJAX request
func IsAjaxRequest(c *fiber.Ctx) bool {
	return c.Get("X-Requested-With") == "XMLHttpRequest"
}

// IsJSONRequest checks if the request content type is JSON
func IsJSONRequest(c *fiber.Ctx) bool {
	contentType := c.Get("Content-Type")
	return contentType == "application/json" || contentType == "application/json; charset=utf-8"
}
