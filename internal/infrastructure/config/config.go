package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	JWT            JWTConfig
	Log            LogConfig
	CORS           CORSConfig
	Metrics        MetricsConfig
	Infrastructure InfrastructureConfig
	OTP            OTPConfig
	Events         EventsConfig
	RateLimiting   RateLimitingConfig
}

// InfrastructureConfig holds infrastructure provider configurations
type InfrastructureConfig struct {
	DatabaseProvider string // postgres, mysql, sqlite, mongodb
	CacheProvider    string // redis, memory, memcached
	StorageProvider  string // s3, local, gcs, azure
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	Host        string
	Environment string
}

// DatabaseConfig holds database configuration with multiple provider support
type DatabaseConfig struct {
	Provider        string // postgres, mysql, sqlite
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	Charset         string
	FilePath        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI             string
	DB              string
	Username        string
	Password        string
	MaxPoolSize     uint64
	MinPoolSize     uint64
	MaxConnIdleTime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	ClusterMode  bool
	ClusterNodes []string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string
	Format     string
	Output     string // stdout, file, syslog
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled     bool
	Provider    string // prometheus, datadog, statsd
	Endpoint    string
	ServiceName string
	Environment string
}

// OTPConfig holds OTP-specific configuration
type OTPConfig struct {
	Expiry         time.Duration
	Length         int
	RedisKeyPrefix string
	CodeCharset    string
}

// EventsConfig holds event system configuration
type EventsConfig struct {
	Enabled       bool
	RedisChannel  string
	BatchSize     int
	FlushInterval time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
	EventTypes    EventTypesConfig
}

// EventTypesConfig holds configuration for different event types
type EventTypesConfig struct {
	OTPGenerated EventTypeConfig
	OTPVerified  EventTypeConfig
	UserCreated  EventTypeConfig
	UserLoggedIn EventTypeConfig
	RateLimited  EventTypeConfig
}

// EventTypeConfig holds configuration for a specific event type
type EventTypeConfig struct {
	Name    string
	Enabled bool
	TTL     time.Duration
}

// RateLimitingConfig holds rate limiting configuration for different endpoints
type RateLimitingConfig struct {
	Global RateLimitConfig
	Auth   RateLimitConfig
	OTP    RateLimitConfig
	User   RateLimitConfig
	Custom map[string]RateLimitConfig
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Duration time.Duration
	Enabled  bool
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		// Silently continue if .env file is not found
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/otp-server")

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, continue with environment variables
	}

	config := &Config{
		Server: ServerConfig{
			Port:        getEnv("SERVER_PORT", "8080"),
			Host:        getEnv("SERVER_HOST", "localhost"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Provider:        getEnv("DB_PROVIDER", "postgres"),
			Host:            getEnv("POSTGRES_HOST", "localhost"),
			Port:            getEnv("POSTGRES_PORT", "5432"),
			User:            getEnv("POSTGRES_USER", "otp_server_user"),
			Password:        getEnv("POSTGRES_PASSWORD", "otp_server_password"),
			DBName:          getEnv("POSTGRES_DB", "otp_server_db"),
			SSLMode:         getEnv("POSTGRES_SSL_MODE", "disable"),
			Charset:         getEnv("MYSQL_CHARSET", "utf8mb4"),
			FilePath:        getEnv("SQLITE_FILE_PATH", "./otp_server.db"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", time.Hour),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 3),
			ClusterMode:  getEnvAsBool("REDIS_CLUSTER_MODE", false),
			ClusterNodes: getEnvAsSlice("REDIS_CLUSTER_NODES", []string{}),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			Expiry:        getEnvAsDuration("JWT_EXPIRY", 24000*time.Hour),
			RefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 42000*time.Hour),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			FilePath:   getEnv("LOG_FILE_PATH", "./logs/app.log"),
			MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvAsInt("LOG_MAX_AGE", 28),
			Compress:   getEnvAsBool("LOG_COMPRESS", true),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With"}),
		},

		Metrics: MetricsConfig{
			Enabled:     getEnvAsBool("METRICS_ENABLED", true),
			Provider:    getEnv("METRICS_PROVIDER", "prometheus"),
			Endpoint:    getEnv("METRICS_ENDPOINT", "/metrics"),
			ServiceName: getEnv("METRICS_SERVICE_NAME", "otp-server"),
			Environment: getEnv("METRICS_ENVIRONMENT", "development"),
		},
		Infrastructure: InfrastructureConfig{
			DatabaseProvider: getEnv("DB_PROVIDER", "postgres"),
			CacheProvider:    getEnv("CACHE_PROVIDER", "redis"),
			StorageProvider:  getEnv("STORAGE_PROVIDER", "s3"),
		},
		OTP: OTPConfig{
			Expiry:         getEnvAsDuration("OTP_EXPIRY", 2*time.Minute),
			Length:         getEnvAsInt("OTP_LENGTH", 6),
			RedisKeyPrefix: getEnv("OTP_REDIS_KEY_PREFIX", "otp"),
			CodeCharset:    getEnv("OTP_CODE_CHARSET", "0123456789"),
		},
		Events: EventsConfig{
			Enabled:       getEnvAsBool("EVENTS_ENABLED", true),
			RedisChannel:  getEnv("EVENTS_REDIS_CHANNEL", "events"),
			BatchSize:     getEnvAsInt("EVENTS_BATCH_SIZE", 100),
			FlushInterval: getEnvAsDuration("EVENTS_FLUSH_INTERVAL", 5*time.Second),
			RetryAttempts: getEnvAsInt("EVENTS_RETRY_ATTEMPTS", 3),
			RetryDelay:    getEnvAsDuration("EVENTS_RETRY_DELAY", time.Second),
			EventTypes: EventTypesConfig{
				OTPGenerated: EventTypeConfig{
					Name:    getEnv("EVENT_OTP_GENERATED_NAME", "otp_generated"),
					Enabled: getEnvAsBool("EVENT_OTP_GENERATED_ENABLED", true),
					TTL:     getEnvAsDuration("EVENT_OTP_GENERATED_TTL", 24*time.Hour),
				},
				OTPVerified: EventTypeConfig{
					Name:    getEnv("EVENT_OTP_VERIFIED_NAME", "otp_verified"),
					Enabled: getEnvAsBool("EVENT_OTP_VERIFIED_ENABLED", true),
					TTL:     getEnvAsDuration("EVENT_OTP_VERIFIED_TTL", 24*time.Hour),
				},
				UserCreated: EventTypeConfig{
					Name:    getEnv("EVENT_USER_CREATED_NAME", "user_created"),
					Enabled: getEnvAsBool("EVENT_USER_CREATED_ENABLED", true),
					TTL:     getEnvAsDuration("EVENT_USER_CREATED_TTL", 7*24*time.Hour),
				},
				UserLoggedIn: EventTypeConfig{
					Name:    getEnv("EVENT_USER_LOGGED_IN_NAME", "user_logged_in"),
					Enabled: getEnvAsBool("EVENT_USER_LOGGED_IN_ENABLED", true),
					TTL:     getEnvAsDuration("EVENT_USER_LOGGED_IN_TTL", 24*time.Hour),
				},
				RateLimited: EventTypeConfig{
					Name:    getEnv("EVENT_RATE_LIMITED_NAME", "rate_limited"),
					Enabled: getEnvAsBool("EVENT_RATE_LIMITED_ENABLED", true),
					TTL:     getEnvAsDuration("EVENT_RATE_LIMITED_TTL", 24*time.Hour),
				},
			},
		},
		RateLimiting: RateLimitingConfig{
			Global: RateLimitConfig{
				Requests: getEnvAsInt("RATE_LIMIT_GLOBAL_REQUESTS", 100),
				Duration: getEnvAsDuration("RATE_LIMIT_GLOBAL_DURATION", time.Minute),
				Enabled:  getEnvAsBool("RATE_LIMIT_GLOBAL_ENABLED", true),
			},
			Auth: RateLimitConfig{
				Requests: getEnvAsInt("RATE_LIMIT_AUTH_REQUESTS", 20),
				Duration: getEnvAsDuration("RATE_LIMIT_AUTH_DURATION", time.Minute),
				Enabled:  getEnvAsBool("RATE_LIMIT_AUTH_ENABLED", true),
			},
			OTP: RateLimitConfig{
				Requests: getEnvAsInt("RATE_LIMIT_OTP_REQUESTS", 3),
				Duration: getEnvAsDuration("RATE_LIMIT_OTP_DURATION", 10*time.Minute),
				Enabled:  getEnvAsBool("RATE_LIMIT_OTP_ENABLED", true),
			},
			User: RateLimitConfig{
				Requests: getEnvAsInt("RATE_LIMIT_USER_REQUESTS", 50),
				Duration: getEnvAsDuration("RATE_LIMIT_USER_DURATION", time.Minute),
				Enabled:  getEnvAsBool("RATE_LIMIT_USER_ENABLED", true),
			},
		},
	}

	return config, nil
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetDatabaseURL returns the database connection string based on provider
func (c *Config) GetDatabaseURL() string {
	switch c.Database.Provider {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			c.Database.User,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.DBName,
			c.Database.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			c.Database.User,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.DBName,
			c.Database.Charset,
		)
	case "sqlite":
		return c.Database.FilePath
	default:
		return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			c.Database.User,
			c.Database.Password,
			c.Database.Host,
			c.Database.Port,
			c.Database.DBName,
			c.Database.SSLMode,
		)
	}
}

// GetRedisURL returns the Redis connection string
func (c *Config) GetRedisURL() string {
	if c.Redis.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%s/%d",
			c.Redis.Password,
			c.Redis.Host,
			c.Redis.Port,
			c.Redis.DB,
		)
	}
	return fmt.Sprintf("redis://%s:%s/%d",
		c.Redis.Host,
		c.Redis.Port,
		c.Redis.DB,
	)
}
