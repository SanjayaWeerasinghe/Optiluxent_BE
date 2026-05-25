package middleware

import (
	"encoding/json"
	"fmt"
	"strings"

	domain "erp-system/internal/domain/audit"
	auditinfra "erp-system/internal/infrastructure/audit"

	"github.com/gofiber/fiber/v2"
)

// AuditLog records all state-changing requests (POST/PUT/PATCH/DELETE) asynchronously.
// Captures who called the API, what was sent (new_values), and when.
// For full old_values on UPDATE/DELETE, call auditLogger.LogAsync explicitly in the service layer.
// Must be placed after the Authenticate middleware so user_id / tenant_id are available.
func AuditLog(auditLogger *auditinfra.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := string(c.Method())
		if !isStateChanging(method) {
			return c.Next()
		}

		// Copy request body before the handler consumes it.
		reqBody := append([]byte(nil), c.Body()...)

		err := c.Next()

		// Only log successful mutations (2xx).
		status := c.Response().StatusCode()
		if status < 200 || status >= 300 {
			return err
		}

		action := methodToAction(method)
		resource := extractResource(c.Path())

		entry := domain.Entry{
			Action:    action,
			Resource:  resource,
			IPAddress: c.IP(),
			UserAgent: c.Get("User-Agent"),
		}

		if uid, ok := c.Locals("user_id").(uint); ok {
			entry.UserID = &uid
		}
		if tid, ok := c.Locals("tenant_id").(uint); ok {
			entry.TenantID = &tid
		}

		// Resource ID: URL param covers PUT/PATCH/DELETE; POST extracts from response body.
		if rid := c.Params("id"); rid != "" {
			entry.ResourceID = rid
		} else if method == "POST" {
			entry.ResourceID = extractIDFromResponse(c.Response().Body())
		}

		// new_values: the JSON payload the caller sent (what was created or updated).
		// DELETE carries no body; old_values for DELETE must be set at the service layer.
		if len(reqBody) > 0 && method != "DELETE" {
			entry.NewValues = json.RawMessage(reqBody)
		}

		auditLogger.LogAsync(c.UserContext(), entry)
		return err
	}
}

// extractIDFromResponse tries to parse {"data":{"id":N}} or {"id":N} from the response.
func extractIDFromResponse(body []byte) string {
	var wrapper struct {
		Data struct {
			ID uint64 `json:"id"`
		} `json:"data"`
		ID uint64 `json:"id"`
	}
	if json.Unmarshal(body, &wrapper) != nil {
		return ""
	}
	if wrapper.Data.ID > 0 {
		return fmt.Sprintf("%d", wrapper.Data.ID)
	}
	if wrapper.ID > 0 {
		return fmt.Sprintf("%d", wrapper.ID)
	}
	return ""
}

func isStateChanging(method string) bool {
	switch method {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	}
	return false
}

func methodToAction(method string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	}
	return strings.ToUpper(method)
}

func extractResource(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for _, p := range parts {
		if p == "api" || p == "v1" || p == "" {
			continue
		}
		return p
	}
	return path
}
