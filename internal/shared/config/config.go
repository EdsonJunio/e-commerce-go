package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contém todas as configurações da aplicação
type Config struct {
	AppName     string
	Version     string
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	CORS        CORSConfig
}

// ServerConfig contém as configurações do servidor HTTP
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig contém as configurações de conexão com o banco de dados
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

// CORSConfig contém as configurações de CORS
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// getEnvString gets an environment variable as a string with a default value
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as an integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration gets an environment variable as a time.Duration with a default value
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

// getEnvStringSlice gets an environment variable as a string slice with a default value
// The environment variable should be a comma-separated list of values
func getEnvStringSlice(key string, defaultValue []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	if value == "" {
		return []string{}
	}

	// Split by comma and trim spaces
	values := strings.Split(value, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}

	return values
}

// getEnvBool gets an environment variable as a boolean with a default value
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

// Load carrega as configurações da aplicação a partir de variáveis de ambiente
func Load() *Config {
	return &Config{
		AppName:     getEnvString("APP_NAME", "e-commerce-go"),
		Version:     getEnvString("APP_VERSION", "1.0.0"),
		Environment: getEnvString("APP_ENVIRONMENT", "development"),

		Server: ServerConfig{
			Port:            getEnvString("SERVER_PORT", "8080"),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
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
		CORS: CORSConfig{
			AllowOrigins:     getEnvStringSlice("CORS_ALLOW_ORIGINS", []string{"*"}),
			AllowMethods:     getEnvStringSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}),
			AllowHeaders:     getEnvStringSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Length", "Content-Type", "Authorization"}),
			AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", false),
			MaxAge:           getEnvDuration("CORS_MAX_AGE", 12*time.Hour),
		},
	}
}
