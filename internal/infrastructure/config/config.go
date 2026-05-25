package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	ClickHouse ClickHouseConfig `mapstructure:"clickhouse"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	App        AppConfig        `mapstructure:"app"`
}

// ClickHouseConfig holds ClickHouse ledger database configuration
type ClickHouseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Database     string `mapstructure:"database"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port              int           `mapstructure:"port"`
	Host              string        `mapstructure:"host"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
	BodyLimit         int           `mapstructure:"body_limit"`
	EnableCORS        bool          `mapstructure:"enable_cors"`
	EnableSwagger     bool          `mapstructure:"enable_swagger"`
	Prefork           bool          `mapstructure:"prefork"`
	EnableCompression bool          `mapstructure:"enable_compression"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	LogLevel        string        `mapstructure:"log_level"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	EnableCache  bool          `mapstructure:"enable_cache"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret               string        `mapstructure:"secret"`
	AccessTokenExpiry    time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry   time.Duration `mapstructure:"refresh_token_expiry"`
	Issuer               string        `mapstructure:"issuer"`
	EnableRefreshToken   bool          `mapstructure:"enable_refresh_token"`
	EnableTokenBlacklist bool          `mapstructure:"enable_token_blacklist"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	MaxAge           int      `mapstructure:"max_age"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json or console
	OutputPath string `mapstructure:"output_path"`
	EnableFile bool   `mapstructure:"enable_file"`
}

// AppConfig holds application-level configuration
type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"` // development, staging, production
	Debug       bool   `mapstructure:"debug"`
	Timezone    string `mapstructure:"timezone"`
}

var cfg *Config

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Override with environment variables
	viper.AutomaticEnv()
	bindEnvVariables()

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Validate configuration
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Get returns the loaded configuration
func Get() *Config {
	if cfg == nil {
		panic("configuration not loaded, call Load() first")
	}
	return cfg
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 30*time.Second)
	viper.SetDefault("server.write_timeout", 30*time.Second)
	viper.SetDefault("server.idle_timeout", 60*time.Second)
	viper.SetDefault("server.shutdown_timeout", 10*time.Second)
	viper.SetDefault("server.body_limit", 4*1024*1024) // 4MB
	viper.SetDefault("server.enable_cors", true)
	viper.SetDefault("server.enable_swagger", true)

	// ClickHouse defaults (ledger DB)
	viper.SetDefault("clickhouse.host", "localhost")
	viper.SetDefault("clickhouse.port", 9000)
	viper.SetDefault("clickhouse.database", "erp_ledger")
	viper.SetDefault("clickhouse.user", "default")
	viper.SetDefault("clickhouse.password", "")
	viper.SetDefault("clickhouse.max_open_conns", 10)
	viper.SetDefault("clickhouse.max_idle_conns", 5)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.database", "erp_db")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", time.Hour)
	viper.SetDefault("database.log_level", "warn")
	viper.SetDefault("database.auto_migrate", false)

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.dial_timeout", 5*time.Second)
	viper.SetDefault("redis.read_timeout", 3*time.Second)
	viper.SetDefault("redis.write_timeout", 3*time.Second)
	viper.SetDefault("redis.enable_cache", true)

	// JWT defaults
	viper.SetDefault("jwt.secret", "change-this-secret-key-in-production")
	viper.SetDefault("jwt.access_token_expiry", 1*time.Hour)
	viper.SetDefault("jwt.refresh_token_expiry", 168*time.Hour) // 7 days
	viper.SetDefault("jwt.issuer", "erp-system")
	viper.SetDefault("jwt.enable_refresh_token", true)
	viper.SetDefault("jwt.enable_token_blacklist", true)

	// CORS defaults
	viper.SetDefault("cors.allow_origins", []string{"*"})
	viper.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	viper.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.expose_headers", []string{"Content-Length"})
	viper.SetDefault("cors.max_age", 86400)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output_path", "stdout")
	viper.SetDefault("logging.enable_file", false)

	// App defaults
	viper.SetDefault("app.name", "ERP System")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", true)
	viper.SetDefault("app.timezone", "UTC")
}

// bindEnvVariables binds environment variables to config keys
func bindEnvVariables() {
	// Server
	_ = viper.BindEnv("server.port", "PORT")
	_ = viper.BindEnv("server.host", "HOST")

	// ClickHouse
	_ = viper.BindEnv("clickhouse.host", "CH_HOST")
	_ = viper.BindEnv("clickhouse.port", "CH_PORT")
	_ = viper.BindEnv("clickhouse.database", "CH_DATABASE")
	_ = viper.BindEnv("clickhouse.user", "CH_USER")
	_ = viper.BindEnv("clickhouse.password", "CH_PASSWORD")

	// Database
	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.database", "DB_NAME")
	_ = viper.BindEnv("database.ssl_mode", "DB_SSL_MODE")

	// Redis
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "REDIS_DB")

	// JWT
	_ = viper.BindEnv("jwt.secret", "JWT_SECRET")

	// App
	_ = viper.BindEnv("app.environment", "ENV", "ENVIRONMENT")
	_ = viper.BindEnv("app.debug", "DEBUG")
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// Validate database config
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate JWT config
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "change-this-secret-key-in-production" {
		if cfg.App.Environment == "production" {
			return fmt.Errorf("JWT secret must be set in production")
		}
	}

	// Validate logging config
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "fatal": true}
	if !validLogLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", cfg.Logging.Level)
	}

	validLogFormats := map[string]bool{"json": true, "console": true}
	if !validLogFormats[cfg.Logging.Format] {
		return fmt.Errorf("invalid log format: %s", cfg.Logging.Format)
	}

	// Validate app config
	validEnvironments := map[string]bool{"development": true, "staging": true, "production": true}
	if !validEnvironments[cfg.App.Environment] {
		return fmt.Errorf("invalid environment: %s", cfg.App.Environment)
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsStaging returns true if running in staging environment
func (c *Config) IsStaging() bool {
	return c.App.Environment == "staging"
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// GetClickHouseDSN returns the ClickHouse native-protocol DSN
func (c *Config) GetClickHouseDSN() string {
	return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
		c.ClickHouse.User,
		c.ClickHouse.Password,
		c.ClickHouse.Host,
		c.ClickHouse.Port,
		c.ClickHouse.Database,
	)
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// LoadFromEnv loads configuration from environment variables only (for testing)
func LoadFromEnv() (*Config, error) {
	// Get environment from ENV variable, default to development
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	cfg = &Config{
		Server: ServerConfig{
			Port:            getEnvAsInt("PORT", 3000),
			Host:            getEnv("HOST", "0.0.0.0"),
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			BodyLimit:       4 * 1024 * 1024,
			EnableCORS:      true,
			EnableSwagger:   true,
		},
		ClickHouse: ClickHouseConfig{
			Host:         getEnv("CH_HOST", "localhost"),
			Port:         getEnvAsInt("CH_PORT", 9000),
			Database:     getEnv("CH_DATABASE", "erp_ledger"),
			User:         getEnv("CH_USER", "default"),
			Password:     getEnv("CH_PASSWORD", ""),
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Database:        getEnv("DB_NAME", "erp_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: time.Hour,
			LogLevel:        "warn",
			AutoMigrate:     false,
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvAsInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     10,
			MinIdleConns: 5,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			EnableCache:  true,
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", "change-this-secret-key-in-production"),
			AccessTokenExpiry:    1 * time.Hour,
			RefreshTokenExpiry:   168 * time.Hour,
			Issuer:               "erp-system",
			EnableRefreshToken:   true,
			EnableTokenBlacklist: true,
		},
		CORS: CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			AllowCredentials: true,
			ExposeHeaders:    []string{"Content-Length"},
			MaxAge:           86400,
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			OutputPath: "stdout",
			EnableFile: false,
		},
		App: AppConfig{
			Name:        "ERP System",
			Version:     "1.0.0",
			Environment: env,
			Debug:       env == "development",
			Timezone:    "UTC",
		},
	}

	return cfg, validate(cfg)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
