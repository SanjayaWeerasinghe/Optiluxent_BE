package events

import (
	"encoding/json"
	"time"
)

// Event is the base structure for all domain events.
type Event struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	TenantID      uint            `json:"tenant_id"`
	UserID        uint            `json:"user_id"`
	CorrelationID string          `json:"correlation_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
}

// EventHandler is a function that processes an incoming event.
type EventHandler func(event Event) error

// Well-known stream names.
const (
	StreamUsers       = "erp:users"
	StreamAuth        = "erp:auth"
	StreamSystem      = "erp:system"
	StreamDLQ         = "erp:dlq"
	StreamProcurement = "erp:procurement"
	StreamSales       = "erp:sales"
)

// Well-known event type names.
const (
	TypeUserCreated      = "user.created"
	TypeUserUpdated      = "user.updated"
	TypeUserDeleted      = "user.deleted"
	TypeUserLogin        = "user.login"
	TypeAuthLoginFailed  = "auth.login_failed"
	TypeAuthLogout       = "auth.logout"
	TypeAuthTokenRefresh = "auth.token_refreshed"

	// Procurement — published for future financial module integration.
	TypeInvoicePosted      = "procurement.invoice.posted"
	TypePaymentRecorded    = "procurement.payment.recorded"

	// Sales — published for future financial module integration.
	TypeSIPosted           = "sales.invoice.posted"
	TypeSIPaymentRecorded  = "sales.payment.recorded"
)
