package postgres

import (
	"context"
	"fmt"
	"time"

	"erp-system/internal/infrastructure/config"
	"erp-system/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

// Connect establishes a connection to PostgreSQL database
func Connect(cfg *config.Config) (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	// Create GORM logger
	gormLog := newGormLogger(cfg.Database.LogLevel)

	// Create GORM config
	gormConfig := &gorm.Config{
		Logger:                                   gormLog,
		DisableForeignKeyConstraintWhenMigrating: false,
		SkipDefaultTransaction:                   true, // Better performance
		PrepareStmt:                              true, // Prepared statement cache
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open connection
	dsn := cfg.GetDSN()
	var err error
	db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Ping database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connected successfully",
		logger.String("host", cfg.Database.Host),
		logger.Int("port", cfg.Database.Port),
		logger.String("database", cfg.Database.Database),
	)

	return db, nil
}

// Get returns the database instance
func Get() *gorm.DB {
	if db == nil {
		panic("database not initialized, call Connect() first")
	}
	return db
}

// Close closes the database connection
func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	logger.Info("Closing database connection")
	return sqlDB.Close()
}

// HealthCheck performs a health check on the database
func HealthCheck(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database statistics
func GetStats() (map[string]interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// Transaction executes a function within a database transaction
func Transaction(fn func(*gorm.DB) error) error {
	return db.Transaction(fn)
}

// WithContext returns a new DB instance with context
func WithContext(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

// gormLogger is a custom logger for GORM
type gormLogger struct {
	level gormlogger.LogLevel
}

// newGormLogger creates a new GORM logger
func newGormLogger(level string) gormlogger.Interface {
	var logLevel gormlogger.LogLevel

	switch level {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Warn
	}

	return &gormLogger{level: logLevel}
}

// LogMode sets the log level
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &gormLogger{level: level}
}

// Info logs info level messages
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn logs warn level messages
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error logs error level messages
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace logs SQL queries
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []logger.Field{
		logger.Duration("duration", elapsed),
		logger.Int64("rows", rows),
	}

	if err != nil && l.level >= gormlogger.Error {
		logger.Error("Database query error",
			append(fields,
				logger.Err(err),
				logger.String("sql", sql),
			)...,
		)
		return
	}

	if elapsed > 200*time.Millisecond && l.level >= gormlogger.Warn {
		logger.Warn("Slow SQL query",
			append(fields, logger.String("sql", sql))...,
		)
		return
	}

	if l.level >= gormlogger.Info {
		logger.Debug("SQL query executed",
			append(fields, logger.String("sql", sql))...,
		)
	}
}
