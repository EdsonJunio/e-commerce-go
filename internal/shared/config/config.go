package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const minimumJWTSecretLength = 32

type Config struct {
	AppName     string
	Version     string
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Security    SecurityConfig
	CORS        CORSConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
	EnablePprof     bool
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        string
}

type RedisConfig struct {
	Host        string
	Port        string
	Password    string
	DB          int
	CategoryTTL time.Duration
}

type JWTConfig struct {
	Secret         string
	Issuer         string
	Audience       string
	AccessTokenTTL time.Duration
}

type SecurityConfig struct {
	LoginMaxAttempts int
	LoginWindow      time.Duration
}

type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load()
	if err := validateEnvironmentValues(); err != nil {
		return nil, err
	}

	cfg := &Config{
		AppName:     getEnvString("APP_NAME", "e-commerce-go"),
		Version:     getEnvString("APP_VERSION", "1.0.0"),
		Environment: strings.ToLower(getEnvString("APP_ENVIRONMENT", "development")),
		Server: ServerConfig{
			Port:            getEnvString("SERVER_PORT", "8081"),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
			MaxBodyBytes:    getEnvInt64("SERVER_MAX_BODY_BYTES", 1<<20),
			EnablePprof:     getEnvBool("ENABLE_PPROF", false),
		},
		Database: DatabaseConfig{
			Host:            getEnvString("DB_HOST", "localhost"),
			Port:            getEnvString("DB_PORT", "5432"),
			User:            getEnvString("DB_USER", "postgres"),
			Password:        getEnvString("DB_PASSWORD", ""),
			Name:            getEnvString("DB_NAME", "ecommerce"),
			SSLMode:         getEnvString("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			LogLevel:        getEnvString("DB_LOG_LEVEL", "warn"),
		},
		Redis: RedisConfig{
			Host:        getEnvString("REDIS_HOST", "localhost"),
			Port:        getEnvString("REDIS_PORT", "6379"),
			Password:    getEnvString("REDIS_PASSWORD", ""),
			DB:          getEnvInt("REDIS_DB", 0),
			CategoryTTL: getEnvDuration("REDIS_CATEGORY_TTL", time.Minute),
		},
		JWT: JWTConfig{
			Secret:         getEnvString("JWT_SECRET", ""),
			Issuer:         getEnvString("JWT_ISSUER", "ecommerce-api"),
			Audience:       getEnvString("JWT_AUDIENCE", "ecommerce-clients"),
			AccessTokenTTL: getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		},
		Security: SecurityConfig{
			LoginMaxAttempts: getEnvInt("LOGIN_RATE_LIMIT_ATTEMPTS", 10),
			LoginWindow:      getEnvDuration("LOGIN_RATE_LIMIT_WINDOW", time.Minute),
		},
		CORS: CORSConfig{
			AllowOrigins:     getEnvStringSlice("CORS_ALLOW_ORIGINS", []string{"http://localhost:3000"}),
			AllowMethods:     getEnvStringSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}),
			AllowHeaders:     getEnvStringSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Length", "Content-Type", "Authorization"}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvDuration("CORS_MAX_AGE", 12*time.Hour),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("configuration is required")
	}
	if c.AppName == "" || c.Version == "" {
		return errors.New("APP_NAME and APP_VERSION are required")
	}
	if c.Environment != "development" && c.Environment != "test" && c.Environment != "production" {
		return fmt.Errorf("unsupported APP_ENVIRONMENT %q", c.Environment)
	}
	if c.Server.Port == "" || c.Server.ReadTimeout <= 0 || c.Server.WriteTimeout <= 0 || c.Server.IdleTimeout <= 0 || c.Server.ShutdownTimeout <= 0 {
		return errors.New("server port and timeouts must be valid")
	}
	if c.Server.MaxBodyBytes <= 0 {
		return errors.New("SERVER_MAX_BODY_BYTES must be positive")
	}
	if c.Database.Host == "" || c.Database.Port == "" || c.Database.User == "" || c.Database.Name == "" {
		return errors.New("database host, port, user, and name are required")
	}
	if c.Database.MaxOpenConns <= 0 || c.Database.MaxIdleConns < 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return errors.New("database pool settings are invalid")
	}
	if c.Redis.Host == "" || c.Redis.Port == "" || c.Redis.CategoryTTL <= 0 {
		return errors.New("Redis host, port, and category TTL must be valid")
	}
	if len(c.JWT.Secret) < minimumJWTSecretLength {
		return fmt.Errorf("JWT_SECRET must contain at least %d characters", minimumJWTSecretLength)
	}
	if c.JWT.Issuer == "" || c.JWT.Audience == "" || c.JWT.AccessTokenTTL <= 0 {
		return errors.New("JWT issuer, audience, and access token TTL must be valid")
	}
	if c.Security.LoginMaxAttempts <= 0 || c.Security.LoginWindow <= 0 {
		return errors.New("login rate limit settings must be valid")
	}
	if len(c.CORS.AllowOrigins) == 0 {
		return errors.New("at least one CORS origin is required")
	}
	if c.Environment == "production" && contains(c.CORS.AllowOrigins, "*") {
		return errors.New("wildcard CORS origins are not allowed in production")
	}
	if c.Environment == "production" && c.Database.Password == "" {
		return errors.New("DB_PASSWORD is required in production")
	}
	return nil
}

func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	if strings.TrimSpace(value) == "" {
		return nil
	}
	values := strings.Split(value, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	return values
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func validateEnvironmentValues() error {
	for _, key := range []string{"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "REDIS_DB", "LOGIN_RATE_LIMIT_ATTEMPTS"} {
		if value, exists := os.LookupEnv(key); exists {
			if _, err := strconv.Atoi(value); err != nil {
				return fmt.Errorf("%s must be an integer: %w", key, err)
			}
		}
	}
	if value, exists := os.LookupEnv("SERVER_MAX_BODY_BYTES"); exists {
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("SERVER_MAX_BODY_BYTES must be an integer: %w", err)
		}
	}
	for _, key := range []string{
		"SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT", "SERVER_IDLE_TIMEOUT", "SERVER_SHUTDOWN_TIMEOUT",
		"DB_CONN_MAX_LIFETIME", "REDIS_CATEGORY_TTL", "JWT_ACCESS_TOKEN_TTL", "LOGIN_RATE_LIMIT_WINDOW", "CORS_MAX_AGE",
	} {
		if value, exists := os.LookupEnv(key); exists {
			if _, err := time.ParseDuration(value); err != nil {
				return fmt.Errorf("%s must be a duration: %w", key, err)
			}
		}
	}
	for _, key := range []string{"ENABLE_PPROF", "CORS_ALLOW_CREDENTIALS"} {
		if value, exists := os.LookupEnv(key); exists {
			if _, err := strconv.ParseBool(value); err != nil {
				return fmt.Errorf("%s must be a boolean: %w", key, err)
			}
		}
	}
	return nil
}
