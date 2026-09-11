package main

import "testing"

func TestLoadSeedConfig(t *testing.T) {
	tests := []struct {
		name         string
		environment  string
		confirmation string
		databaseURL  string
		adminEmail   string
		adminPass    string
		userPass     string
		wantErr      bool
	}{
		{name: "valid", environment: "development", confirmation: "truncate-ecommerce", databaseURL: "postgres://local", adminEmail: "admin@example.com", adminPass: "admin-password", userPass: "customer-password"},
		{name: "wrong environment", environment: "production", confirmation: "truncate-ecommerce", databaseURL: "postgres://prod", adminEmail: "admin@example.com", adminPass: "admin-password", userPass: "customer-password", wantErr: true},
		{name: "missing confirmation", environment: "development", databaseURL: "postgres://local", adminEmail: "admin@example.com", adminPass: "admin-password", userPass: "customer-password", wantErr: true},
		{name: "missing database URL", environment: "development", confirmation: "truncate-ecommerce", adminEmail: "admin@example.com", adminPass: "admin-password", userPass: "customer-password", wantErr: true},
		{name: "short password", environment: "development", confirmation: "truncate-ecommerce", databaseURL: "postgres://local", adminEmail: "admin@example.com", adminPass: "short", userPass: "customer-password", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENVIRONMENT", tt.environment)
			t.Setenv("SEED_CONFIRM", tt.confirmation)
			t.Setenv("SEED_DATABASE_URL", tt.databaseURL)
			t.Setenv("SEED_ADMIN_EMAIL", tt.adminEmail)
			t.Setenv("SEED_ADMIN_PASSWORD", tt.adminPass)
			t.Setenv("SEED_USER_PASSWORD", tt.userPass)

			got, err := loadSeedConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("loadSeedConfig() error = %v, want error %v", err, tt.wantErr)
			}
			if !tt.wantErr && (got.databaseURL != tt.databaseURL || got.adminEmail != tt.adminEmail) {
				t.Fatalf("loadSeedConfig() = %+v", got)
			}
		})
	}
}
