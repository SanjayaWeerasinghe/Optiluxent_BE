package audit

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	domain "erp-system/internal/domain/audit"
)

type ClickHouseRepository struct {
	db *sql.DB
}

func NewClickHouseRepository(db *sql.DB) domain.Repository {
	return &ClickHouseRepository{db: db}
}

func (r *ClickHouseRepository) Create(ctx context.Context, log *domain.Log) error {
	if r.db == nil {
		return nil
	}
	tenantID := uint64(0)
	if log.TenantID != nil {
		tenantID = uint64(*log.TenantID)
	}
	userID := uint64(0)
	if log.UserID != nil {
		userID = uint64(*log.UserID)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_logs
			(tenant_id, user_id, action, resource, resource_id,
			 old_values, new_values, ip_address, user_agent, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tenantID, userID,
		log.Action, log.Resource, log.ResourceID,
		string(log.OldValues), string(log.NewValues),
		log.IPAddress, log.UserAgent,
		log.CreatedAt,
	)
	return err
}

// GetByID is not meaningful for ClickHouse's append-only ledger; returns nil.
func (r *ClickHouseRepository) GetByID(_ context.Context, _ uint) (*domain.Log, error) {
	return nil, nil
}

func (r *ClickHouseRepository) List(ctx context.Context, f domain.ListFilter) ([]*domain.Log, int64, error) {
	if r.db == nil {
		return nil, 0, nil
	}
	conds := []string{}
	args := []any{}

	if f.TenantID != nil {
		conds = append(conds, "tenant_id = ?")
		args = append(args, uint64(*f.TenantID))
	}
	if f.UserID != nil {
		conds = append(conds, "user_id = ?")
		args = append(args, uint64(*f.UserID))
	}
	if f.Resource != "" {
		conds = append(conds, "resource = ?")
		args = append(args, f.Resource)
	}
	if f.Action != "" {
		conds = append(conds, "action = ?")
		args = append(args, f.Action)
	}
	if f.DateFrom != "" {
		conds = append(conds, "created_at >= ?")
		args = append(args, f.DateFrom)
	}
	if f.DateTo != "" {
		conds = append(conds, "created_at <= ?")
		args = append(args, f.DateTo)
	}

	where := "1=1"
	if len(conds) > 0 {
		where = strings.Join(conds, " AND ")
	}

	var total int64
	countArgs := append([]any{}, args...)
	_ = r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT count() FROM audit_logs WHERE %s", where),
		countArgs...,
	).Scan(&total)

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(`
		SELECT tenant_id, user_id, action, resource, resource_id,
		       old_values, new_values, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, where)

	args = append(args, uint64(limit), uint64(f.Offset))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*domain.Log
	for rows.Next() {
		var (
			tenantID, userID uint64
			oldVal, newVal   string
			entry            domain.Log
		)
		if err := rows.Scan(
			&tenantID, &userID,
			&entry.Action, &entry.Resource, &entry.ResourceID,
			&oldVal, &newVal,
			&entry.IPAddress, &entry.UserAgent,
			&entry.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if tenantID > 0 {
			tid := uint(tenantID)
			entry.TenantID = &tid
		}
		if userID > 0 {
			uid := uint(userID)
			entry.UserID = &uid
		}
		entry.OldValues = []byte(oldVal)
		entry.NewValues = []byte(newVal)
		logs = append(logs, &entry)
	}
	return logs, total, rows.Err()
}
