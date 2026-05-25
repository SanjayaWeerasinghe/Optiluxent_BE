package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface defines logging methods
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger

	Sync() error
}

// Field is an alias for zap.Field
type Field = zap.Field

// zapLogger wraps zap.Logger to implement our Logger interface
type zapLogger struct {
	logger *zap.Logger
}

var (
	// global logger instance
	global Logger
)

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error, fatal
	Format     string // json or console
	OutputPath string // stdout, stderr, or file path
	EnableFile bool
}

// Initialize initializes the global logger
func Initialize(cfg Config) error {
	logger, err := New(cfg)
	if err != nil {
		return err
	}
	global = logger
	return nil
}

// New creates a new logger instance
func New(cfg Config) (Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Choose encoder
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create output paths
	var outputs []zapcore.WriteSyncer
	if cfg.OutputPath == "stdout" || cfg.OutputPath == "" {
		outputs = append(outputs, zapcore.AddSync(os.Stdout))
	} else if cfg.OutputPath == "stderr" {
		outputs = append(outputs, zapcore.AddSync(os.Stderr))
	}

	// Add file output if enabled
	if cfg.EnableFile {
		file, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, zapcore.AddSync(file))
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(outputs...),
		level,
	)

	// Create logger
	zapLog := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &zapLogger{logger: zapLog}, nil
}

// Get returns the global logger instance
func Get() Logger {
	if global == nil {
		// Create default logger if not initialized
		logger, _ := New(Config{
			Level:      "info",
			Format:     "console",
			OutputPath: "stdout",
		})
		global = logger
	}
	return global
}

// Debug logs a debug message
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs an error message
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// WithContext creates a logger with context fields
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	fields := extractContextFields(ctx)
	return l.With(fields...)
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// extractContextFields extracts common fields from context
func extractContextFields(ctx context.Context) []Field {
	var fields []Field

	// Extract request ID if present
	if reqID, ok := ctx.Value("request_id").(string); ok {
		fields = append(fields, String("request_id", reqID))
	}

	// Extract user ID if present
	if userID, ok := ctx.Value("user_id").(uint); ok {
		fields = append(fields, Uint("user_id", userID))
	}

	// Extract tenant ID if present
	if tenantID, ok := ctx.Value("tenant_id").(uint); ok {
		fields = append(fields, Uint("tenant_id", tenantID))
	}

	return fields
}

// Global logging functions that use the global logger

// Debug logs a debug message using global logger
func Debug(msg string, fields ...Field) {
	Get().Debug(msg, fields...)
}

// Info logs an info message using global logger
func Info(msg string, fields ...Field) {
	Get().Info(msg, fields...)
}

// Warn logs a warning message using global logger
func Warn(msg string, fields ...Field) {
	Get().Warn(msg, fields...)
}

// Error logs an error message using global logger
func Error(msg string, fields ...Field) {
	Get().Error(msg, fields...)
}

// Fatal logs a fatal message using global logger and exits
func Fatal(msg string, fields ...Field) {
	Get().Fatal(msg, fields...)
}

// With creates a child logger with additional fields using global logger
func With(fields ...Field) Logger {
	return Get().With(fields...)
}

// WithContext creates a logger with context fields using global logger
func WithContext(ctx context.Context) Logger {
	return Get().WithContext(ctx)
}

// Sync flushes any buffered log entries from global logger
func Sync() error {
	return Get().Sync()
}

// Helper functions for creating fields

// String creates a string field
func String(key, val string) Field {
	return zap.String(key, val)
}

// Int creates an int field
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Int64 creates an int64 field
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Uint creates a uint field
func Uint(key string, val uint) Field {
	return zap.Uint(key, val)
}

// Uint64 creates a uint64 field
func Uint64(key string, val uint64) Field {
	return zap.Uint64(key, val)
}

// Float64 creates a float64 field
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Bool creates a bool field
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Time creates a time field
func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}

// Duration creates a duration field
func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

// Error creates an error field
func Err(err error) Field {
	return zap.Error(err)
}

// Any creates a field with any type
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Stack creates a stack trace field
func Stack() Field {
	return zap.Stack("stacktrace")
}
