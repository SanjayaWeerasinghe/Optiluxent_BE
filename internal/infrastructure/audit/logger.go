package audit

import (
	"context"
	"encoding/json"
	"time"

	domain "erp-system/internal/domain/audit"
	"erp-system/pkg/logger"
)

// Logger writes audit entries to the database.
type Logger struct {
	repo domain.Repository
}

func NewLogger(repo domain.Repository) *Logger {
	return &Logger{repo: repo}
}

func (l *Logger) Log(ctx context.Context, entry domain.Entry) error {
	log := buildLog(entry)
	return l.repo.Create(ctx, log)
}

// LogAsync writes the audit entry in a background goroutine (fire-and-forget).
func (l *Logger) LogAsync(ctx context.Context, entry domain.Entry) {
	go func() {
		if err := l.Log(context.Background(), entry); err != nil {
			logger.Error("audit: failed to write log entry",
				logger.Err(err),
				logger.String("action", entry.Action),
				logger.String("resource", entry.Resource),
			)
		}
	}()
}

func buildLog(entry domain.Entry) *domain.Log {
	log := &domain.Log{
		TenantID:   entry.TenantID,
		UserID:     entry.UserID,
		Action:     entry.Action,
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		IPAddress:  entry.IPAddress,
		UserAgent:  entry.UserAgent,
		CreatedAt:  time.Now(),
	}

	if entry.OldValues != nil {
		if b, err := json.Marshal(entry.OldValues); err == nil {
			log.OldValues = b
		}
	}
	if entry.NewValues != nil {
		if b, err := json.Marshal(entry.NewValues); err == nil {
			log.NewValues = b
		}
	}

	return log
}
