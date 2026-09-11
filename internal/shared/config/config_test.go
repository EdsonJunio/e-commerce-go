package config

import (
	"strings"
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	base := func() *Config {
		return &Config{
			AppName:     "e-commerce-go",
			Version:     "test",
			Environment: "test",
			Server: ServerConfig{
				Port: "8081", ReadTimeout: time.Second, WriteTimeout: time.Second,
				IdleTimeout: time.Second, ShutdownTimeout: time.Second, MaxBodyBytes: 1024,
			},
			Database: DatabaseConfig{
				Host: "localhost", Port: "5432", User: "postgres", Name: "test",
				MaxOpenConns: 5, MaxIdleConns: 1,
			},
			Redis: RedisConfig{Host: "localhost", Port: "6379", CategoryTTL: time.Minute},
			JWT: JWTConfig{
				Secret: strings.Repeat("a", 32), Issuer: "issuer", Audience: "audience", AccessTokenTTL: time.Minute,
			},
			Security: SecurityConfig{LoginMaxAttempts: 10, LoginWindow: time.Minute},
			CORS:     CORSConfig{AllowOrigins: []string{"http://localhost:3000"}},
		}
	}

	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{name: "valid", mutate: func(*Config) {}},
		{name: "short JWT secret", mutate: func(c *Config) { c.JWT.Secret = "short" }, wantErr: true},
		{name: "invalid environment", mutate: func(c *Config) { c.Environment = "staging" }, wantErr: true},
		{name: "unsafe production CORS", mutate: func(c *Config) {
			c.Environment = "production"
			c.CORS.AllowOrigins = []string{"*"}
		}, wantErr: true},
		{name: "missing production database password", mutate: func(c *Config) {
			c.Environment = "production"
			c.Database.Password = ""
		}, wantErr: true},
		{name: "invalid pool", mutate: func(c *Config) { c.Database.MaxIdleConns = 6 }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base()
			tt.mutate(cfg)
			if gotErr := cfg.Validate() != nil; gotErr != tt.wantErr {
				t.Fatalf("Validate() error = %v, want error %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestLoadReadsJWTConfiguration(t *testing.T) {
	t.Setenv("APP_ENVIRONMENT", "test")
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("JWT_ISSUER", "test-issuer")
	t.Setenv("JWT_AUDIENCE", "test-audience")
	t.Setenv("JWT_ACCESS_TOKEN_TTL", "10m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.JWT.Issuer != "test-issuer" || cfg.JWT.Audience != "test-audience" || cfg.JWT.AccessTokenTTL != 10*time.Minute {
		t.Fatalf("unexpected JWT configuration: %+v", cfg.JWT)
	}
}

func TestLoadRejectsMalformedEnvironmentValue(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("SERVER_READ_TIMEOUT", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("Load() expected an error")
	}
}
