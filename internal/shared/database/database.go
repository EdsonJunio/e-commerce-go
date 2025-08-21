package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"e-commerce-go/internal/shared/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// ConnectDB estabelece uma conexão com o banco de dados usando as configurações fornecidas
func ConnectDB() (*gorm.DB, error) {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	masked := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	fmt.Fprintf(os.Stdout, "[db] Attempting to connect to database: %s\n", masked)

	gormCfg := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 newGormLogger(cfg.Database.LogLevel),
	}

	var (
		db    *gorm.DB
		sqlDB *sql.DB
		err   error
	)

	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context canceled while connecting to database")
		default:
			db, err = gorm.Open(postgres.Open(dsn), gormCfg)
			if err == nil {
				sqlDB, err = db.DB()
				if err != nil {
					return nil, fmt.Errorf("failed to get database instance: %w", err)
				}

				configurePool(sqlDB)

				if err = sqlDB.PingContext(ctx); err == nil {
					fmt.Fprintf(os.Stdout, "[db] Successfully connected to database after %d attempts\n", i+1)
					return db, nil
				}
			}

			if i < maxRetries-1 {
				nextRetry := time.Duration(i+1) * retryDelay
				fmt.Fprintf(os.Stdout, "[db] Attempt %d/%d failed: %v. Retrying in %v...\n",
					i+1, maxRetries, err, nextRetry)
				time.Sleep(nextRetry)
			}
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// configurePool configura o pool de conexões do banco de dados
func configurePool(sqlDB *sql.DB) {
	cfg := config.Load()

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	log.Printf("Pool de conexões configurado: MaxIdleConns=%d, MaxOpenConns=%d, ConnMaxLifetime=%v",
		cfg.Database.MaxIdleConns, cfg.Database.MaxOpenConns, cfg.Database.ConnMaxLifetime)
}

// newGormLogger cria um novo logger para o GORM baseado no nível especificado
func newGormLogger(level string) glogger.Interface {
	var lvl glogger.LogLevel

	switch level {
	case "silent":
		lvl = glogger.Silent
	case "error":
		lvl = glogger.Error
	case "warn":
		lvl = glogger.Warn
	case "info":
		lvl = glogger.Info
	default:
		lvl = glogger.Warn
	}

	log.Printf("Nível de log do GORM configurado para: %s", level)
	return glogger.Default.LogMode(lvl)
}
