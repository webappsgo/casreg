package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Storage   StorageConfig
	Security  SecurityConfig
	SMTP      SMTPConfig
	Redis     RedisConfig
	Features  FeaturesConfig
	RateLimit RateLimitConfig
	Quota     QuotaConfig
	LogLevel  string
}

type ServerConfig struct {
	Host           string
	Port           int
	BaseURL        string
	Debug          bool
	MaxRequestSize int64
	RequestTimeout time.Duration
}

type DatabaseConfig struct {
	Type        string // sqlite, postgres, mysql, sqlserver, mongodb
	URL         string
	Host        string
	Port        int
	Name        string
	User        string
	Password    string
	SSLMode     string
	PoolSize    int
	Migrations  bool
}

type StorageConfig struct {
	Backend      string // local, s3, nfs
	Path         string
	S3Endpoint   string
	S3Bucket     string
	S3Region     string
	S3AccessKey  string
	S3SecretKey  string
	S3UseSSL     bool
	NFSServer    string
	NFSPath      string
}

type SecurityConfig struct {
	JWTSecret         string
	JWTExpiration     time.Duration
	BcryptCost        int
	TrustedIPs        []string
	PasskeysEnabled   bool
}

type SMTPConfig struct {
	Enabled        bool
	Host           string
	Port           int
	Username       string
	Password       string
	From           string
	UseTLS         bool
	UseSTARTTLS    bool
	SubjectPrefix  string
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     int
	Password string
	Database int
	SSL      bool
	PoolSize int
}

type FeaturesConfig struct {
	SearchEnabled            bool
	QuotasEnabled            bool
	NotificationsEnabled     bool
	ScanningEnabled          bool
	SignatureVerification    bool
	AllowPublicRegistries    bool
	AllowGuestAccess         bool
	AllowUserRegistration    bool
}

type RateLimitConfig struct {
	Enabled         bool
	Requests        int
	Window          time.Duration
	AnonymousLimit  int
}

type QuotaConfig struct {
	DefaultUserQuota   int64
	DefaultOrgQuota    int64
	DefaultRepoQuota   int64
	GracePeriod        time.Duration
	CleanupInterval    time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:           getEnv("CASREG_HOST", "0.0.0.0"),
			Port:           getEnvInt("CASREG_PORT", 8080),
			BaseURL:        getEnv("CASREG_BASE_URL", "http://localhost:8080"),
			Debug:          getEnvBool("CASREG_DEBUG", false),
			MaxRequestSize: getEnvInt64("CASREG_MAX_REQUEST_SIZE", 100*1024*1024), // 100MB
			RequestTimeout: getEnvDuration("CASREG_REQUEST_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Type:       getEnv("CASREG_DATABASE_TYPE", "sqlite"),
			URL:        getEnv("CASREG_DATABASE_URL", ""),
			Host:       getEnv("CASREG_DATABASE_HOST", ""),
			Port:       getEnvInt("CASREG_DATABASE_PORT", 5432),
			Name:       getEnv("CASREG_DATABASE_NAME", "casreg"),
			User:       getEnv("CASREG_DATABASE_USER", ""),
			Password:   getEnv("CASREG_DATABASE_PASSWORD", ""),
			SSLMode:    getEnv("CASREG_DATABASE_SSL_MODE", "disable"),
			PoolSize:   getEnvInt("CASREG_DATABASE_POOL_SIZE", 25),
			Migrations: getEnvBool("CASREG_DATABASE_MIGRATIONS", true),
		},
		Storage: StorageConfig{
			Backend:     getEnv("CASREG_STORAGE_BACKEND", "local"),
			Path:        getEnv("CASREG_STORAGE_PATH", "/var/lib/casreg/storage"),
			S3Endpoint:  getEnv("CASREG_STORAGE_S3_ENDPOINT", ""),
			S3Bucket:    getEnv("CASREG_STORAGE_S3_BUCKET", ""),
			S3Region:    getEnv("CASREG_STORAGE_S3_REGION", "us-east-1"),
			S3AccessKey: getEnv("CASREG_STORAGE_S3_ACCESS_KEY", ""),
			S3SecretKey: getEnv("CASREG_STORAGE_S3_SECRET_KEY", ""),
			S3UseSSL:    getEnvBool("CASREG_STORAGE_S3_USE_SSL", true),
			NFSServer:   getEnv("CASREG_STORAGE_NFS_SERVER", ""),
			NFSPath:     getEnv("CASREG_STORAGE_NFS_PATH", ""),
		},
		Security: SecurityConfig{
			JWTSecret:       getEnv("CASREG_JWT_SECRET", generateRandomSecret()),
			JWTExpiration:   getEnvDuration("CASREG_JWT_EXPIRATION", 24*time.Hour),
			BcryptCost:      getEnvInt("CASREG_BCRYPT_COST", 12),
			TrustedIPs:      getEnvSlice("CASREG_TRUSTED_IPS", []string{}),
			PasskeysEnabled: getEnvBool("CASREG_PASSKEYS_ENABLED", true),
		},
		SMTP: SMTPConfig{
			Enabled:       getEnvBool("CASREG_SMTP_ENABLED", false),
			Host:          getEnv("CASREG_SMTP_HOST", ""),
			Port:          getEnvInt("CASREG_SMTP_PORT", 587),
			Username:      getEnv("CASREG_SMTP_USERNAME", ""),
			Password:      getEnv("CASREG_SMTP_PASSWORD", ""),
			From:          getEnv("CASREG_SMTP_FROM", ""),
			UseTLS:        getEnvBool("CASREG_SMTP_USE_TLS", true),
			UseSTARTTLS:   getEnvBool("CASREG_SMTP_USE_STARTTLS", false),
			SubjectPrefix: getEnv("CASREG_SMTP_SUBJECT_PREFIX", "[casreg]"),
		},
		Redis: RedisConfig{
			Enabled:  getEnvBool("CASREG_REDIS_ENABLED", false),
			Host:     getEnv("CASREG_REDIS_HOST", "localhost"),
			Port:     getEnvInt("CASREG_REDIS_PORT", 6379),
			Password: getEnv("CASREG_REDIS_PASSWORD", ""),
			Database: getEnvInt("CASREG_REDIS_DATABASE", 0),
			SSL:      getEnvBool("CASREG_REDIS_SSL", false),
			PoolSize: getEnvInt("CASREG_REDIS_POOL_SIZE", 10),
		},
		Features: FeaturesConfig{
			SearchEnabled:         getEnvBool("CASREG_SEARCH_ENABLED", true),
			QuotasEnabled:         getEnvBool("CASREG_QUOTAS_ENABLED", true),
			NotificationsEnabled:  getEnvBool("CASREG_NOTIFICATIONS_ENABLED", true),
			ScanningEnabled:       getEnvBool("CASREG_SCANNING_ENABLED", true),
			SignatureVerification: getEnvBool("CASREG_SIGNATURE_VERIFICATION", true),
			AllowPublicRegistries: getEnvBool("CASREG_ALLOW_PUBLIC_REGISTRIES", true),
			AllowGuestAccess:      getEnvBool("CASREG_ALLOW_GUEST_ACCESS", true),
			AllowUserRegistration: getEnvBool("CASREG_ALLOW_USER_REGISTRATION", true),
		},
		RateLimit: RateLimitConfig{
			Enabled:        getEnvBool("CASREG_RATE_LIMIT_ENABLED", true),
			Requests:       getEnvInt("CASREG_RATE_LIMIT_REQUESTS", 1000),
			Window:         getEnvDuration("CASREG_RATE_LIMIT_WINDOW", 1*time.Hour),
			AnonymousLimit: getEnvInt("CASREG_RATE_LIMIT_ANONYMOUS", 100),
		},
		Quota: QuotaConfig{
			DefaultUserQuota:  getEnvInt64("CASREG_DEFAULT_USER_QUOTA", 0), // 0 = unlimited
			DefaultOrgQuota:   getEnvInt64("CASREG_DEFAULT_ORG_QUOTA", 0),
			DefaultRepoQuota:  getEnvInt64("CASREG_DEFAULT_REPO_QUOTA", 0),
			GracePeriod:       getEnvDuration("CASREG_QUOTA_GRACE_PERIOD", 24*time.Hour),
			CleanupInterval:   getEnvDuration("CASREG_QUOTA_CLEANUP_INTERVAL", 1*time.Hour),
		},
		LogLevel: getEnv("CASREG_LOG_LEVEL", "info"),
	}

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		value = strings.ToLower(value)
		return value == "true" || value == "yes" || value == "1" || value == "enable" || value == "on"
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func generateRandomSecret() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	secret := make([]byte, 64)
	for i := range secret {
		secret[i] = chars[i%len(chars)]
	}
	return string(secret)
}

// Validate checks that the loaded configuration is consistent and complete.
func Validate(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if cfg.Security.BcryptCost < 10 || cfg.Security.BcryptCost > 31 {
		return fmt.Errorf("bcrypt cost must be between 10 and 31, got %d", cfg.Security.BcryptCost)
	}
	if cfg.SMTP.Enabled && cfg.SMTP.Host == "" {
		return fmt.Errorf("SMTP enabled but CASREG_SMTP_HOST is not set")
	}
	if cfg.SMTP.Enabled && cfg.SMTP.From == "" {
		return fmt.Errorf("SMTP enabled but CASREG_SMTP_FROM is not set")
	}
	validDatabaseTypes := map[string]bool{
		"sqlite": true, "postgres": true, "mysql": true, "sqlserver": true, "mongodb": true,
	}
	if !validDatabaseTypes[cfg.Database.Type] {
		return fmt.Errorf("unsupported database type %q; must be one of: sqlite, postgres, mysql, sqlserver, mongodb", cfg.Database.Type)
	}
	validStorageBackends := map[string]bool{
		"local": true, "s3": true, "nfs": true,
	}
	if !validStorageBackends[cfg.Storage.Backend] {
		return fmt.Errorf("unsupported storage backend %q; must be one of: local, s3, nfs", cfg.Storage.Backend)
	}
	if cfg.Storage.Backend == "s3" && cfg.Storage.S3Bucket == "" {
		return fmt.Errorf("S3 storage selected but CASREG_STORAGE_S3_BUCKET is not set")
	}
	if cfg.Storage.Backend == "nfs" && cfg.Storage.NFSServer == "" {
		return fmt.Errorf("NFS storage selected but CASREG_STORAGE_NFS_SERVER is not set")
	}
	return nil
}
