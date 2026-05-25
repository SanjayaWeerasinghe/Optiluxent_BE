package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"

	"erp-system/internal/infrastructure/config"
	"erp-system/pkg/logger"
)

var db *sql.DB

// Connect opens a connection pool to ClickHouse using the native protocol.
func Connect(cfg *config.Config) (*sql.DB, error) {
	if db != nil {
		return db, nil
	}

	db = ch.OpenDB(&ch.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
		Auth: ch.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.User,
			Password: cfg.ClickHouse.Password,
		},
		Settings: ch.Settings{
			"max_execution_time": 60,
		},
		Compression: &ch.Compression{
			Method: ch.CompressionLZ4,
		},
		DialTimeout: 5 * time.Second,
	})

	db.SetMaxOpenConns(cfg.ClickHouse.MaxOpenConns)
	db.SetMaxIdleConns(cfg.ClickHouse.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	logger.Info("ClickHouse ledger DB connected",
		logger.String("host", cfg.ClickHouse.Host),
		logger.Int("port", cfg.ClickHouse.Port),
		logger.String("database", cfg.ClickHouse.Database),
	)

	return db, nil
}

// Close closes the ClickHouse connection pool.
func Close() error {
	if db == nil {
		return nil
	}
	logger.Info("Closing ClickHouse connection")
	return db.Close()
}

// HealthCheck pings ClickHouse.
func HealthCheck(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("ClickHouse not initialised")
	}
	return db.PingContext(ctx)
}

// Migrate runs CREATE TABLE IF NOT EXISTS for all ledger tables.
// Uses LowCardinality for repeated string columns to save memory and speed up filters.
func Migrate(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("ClickHouse not initialised")
	}
	for _, stmt := range schemaDDL {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("ClickHouse schema migration failed: %w", err)
		}
	}
	logger.Info("ClickHouse ledger schema ready")
	return nil
}

// schemaDDL contains the table definitions for the ledger database.
// MergeTree engine: append-only, partitioned by month for efficient pruning.
// LowCardinality reduces storage for repeated strings (action, resource, transaction_type).
var schemaDDL = []string{
	`CREATE TABLE IF NOT EXISTS audit_logs (
		tenant_id   UInt64,
		user_id     UInt64,
		action      LowCardinality(String),
		resource    LowCardinality(String),
		resource_id String,
		old_values  String,
		new_values  String,
		ip_address  String,
		user_agent  String,
		created_at  DateTime64(3) DEFAULT now64()
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMM(created_at)
	ORDER BY (tenant_id, created_at)
	TTL toDateTime(created_at) + INTERVAL 3 YEAR`,

	`CREATE TABLE IF NOT EXISTS stock_ledger (
		tenant_id        UInt64,
		product_id       UInt64,
		variant_id       UInt64,
		warehouse_id     UInt64,
		location_id      UInt64,
		transaction_type LowCardinality(String),
		reference_type   LowCardinality(String),
		reference_id     UInt64,
		quantity         Decimal(18, 4),
		unit_cost        Decimal(18, 4),
		total_cost       Decimal(18, 4),
		transaction_date Date,
		notes            String,
		created_at       DateTime64(3) DEFAULT now64(),
		created_by       UInt64
	) ENGINE = MergeTree()
	PARTITION BY toYYYYMM(transaction_date)
	ORDER BY (tenant_id, transaction_date, product_id)`,
}
