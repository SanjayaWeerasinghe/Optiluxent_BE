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

		action := deriveAction(method, c.Path())
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

// deriveAction turns the (method, path) pair into a specific action name so
// the audit log reads distinctly per operation instead of blurring every
// mutation into CREATE / UPDATE / DELETE. Examples:
//
//   POST   /purchase-orders               → CREATE
//   PUT    /purchase-orders/1             → UPDATE
//   DELETE /purchase-orders/1             → DELETE
//   POST   /purchase-orders/1/confirm     → CONFIRM
//   POST   /purchase-orders/1/cancel      → CANCEL
//   POST   /purchase-orders/1/submit      → SUBMIT
//   POST   /purchase-orders/1/approve     → APPROVE
//   POST   /purchase-orders/1/reject      → REJECT
//   POST   /purchase-orders/1/pay         → PAY
//   POST   /goods-receipts/1/items        → LINE_ADDED
//   PUT    /goods-receipts/1/items/2      → LINE_UPDATED
//   DELETE /goods-receipts/1/items/2      → LINE_DELETED
//   POST   /goods-receipts/1/lines        → LINE_ADDED   (aliased sub-resource)
//   POST   /manufacturing/orders/1/start  → START
func deriveAction(method, path string) string {
	// Walk the path segments and note any known sub-resource / verb after the
	// first numeric :id segment. Anything before :id is module scaffolding.
	segs := strings.Split(strings.TrimPrefix(path, "/"), "/")
	postID := false
	for _, s := range segs {
		if !postID {
			if isNumericSegment(s) {
				postID = true
			}
			continue
		}
		// First segment after the primary :id is the sub-resource / verb.
		switch s {
		case "items", "lines", "outputs", "resources", "fields":
			// Line-level sub-resource CRUD
			switch method {
			case "POST":
				return "LINE_ADDED"
			case "PUT", "PATCH":
				return "LINE_UPDATED"
			case "DELETE":
				return "LINE_DELETED"
			}
		case "confirm", "cancel", "submit", "approve", "reject",
			"start", "complete", "release", "send", "receive",
			"post", "pay", "accept", "finalize":
			return strings.ToUpper(s)
		}
		break
	}
	return methodToAction(method)
}

func isNumericSegment(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// extractResource narrows the path down to the specific doc-kind so audit
// filtering can distinguish e.g. purchase-orders from goods-receipts within
// the same procurement module.
//
// /api/v1/procurement/purchase-orders/1/confirm → "purchase-orders"
// /api/v1/inventory/material-requests/42       → "material-requests"
// /api/v1/masterdata/products/                 → "products"
// /api/v1/auth/login                           → "auth" (single-segment paths fall back)
func extractResource(path string) string {
	segments := make([]string, 0, 4)
	for _, p := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if p == "api" || p == "v1" || p == "" {
			continue
		}
		segments = append(segments, p)
		if len(segments) >= 2 {
			break
		}
	}
	// Prefer the doc-kind (second segment) when the module has one; fall back
	// to the first segment for flat paths like /auth/login.
	if len(segments) >= 2 {
		return segments[1]
	}
	if len(segments) == 1 {
		return segments[0]
	}
	return path
}
