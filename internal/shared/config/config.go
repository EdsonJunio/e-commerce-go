package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config contém todas as configurações da aplicação
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	CORS     CORSConfig
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

// Load carrega as configurações da aplicação a partir de variáveis de ambiente
func Load() *Config {
	return &Config{
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

// getEnvString retorna uma string do ambiente ou um valor padrão
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt retorna um inteiro do ambiente ou um valor padrão
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool retorna um booleano do ambiente ou um valor padrão
func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvDuration retorna uma duração do ambiente ou um valor padrão
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvStringSlice retorna um slice de strings do ambiente ou um valor padrão
func getEnvStringSlice(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}
