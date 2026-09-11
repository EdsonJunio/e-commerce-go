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

func ConnectDB(ctx context.Context, cfg config.DatabaseConfig) (*gorm.DB, error) {
	host := cfg.Host

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)

	masked := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s sslmode=%s",
		host,
		cfg.Port,
		cfg.User,
		cfg.Name,
		cfg.SSLMode,
	)
	fmt.Fprintf(os.Stdout, "[db] Attempting to connect to database: %s\n", masked)

	gormCfg := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 newGormLogger(cfg.LogLevel),
	}

	var (
		db    *gorm.DB
		sqlDB *sql.DB
		err   error
	)

	const maxRetries = 8
	retryDelay := 250 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("database connection canceled: %w", ctx.Err())
		default:
			db, err = gorm.Open(postgres.Open(dsn), gormCfg)
			if err == nil {
				sqlDB, err = db.DB()
				if err != nil {
					return nil, fmt.Errorf("failed to get database instance: %w", err)
				}

				configurePool(sqlDB, cfg)

				if err = sqlDB.PingContext(ctx); err == nil {
					fmt.Fprintf(
						os.Stdout,
						"[db] Successfully connected to database after %d attempts\n",
						i+1,
					)
					return db, nil
				}
				_ = sqlDB.Close()
			}

			if i < maxRetries-1 {
				fmt.Fprintf(
					os.Stdout,
					"[db] Attempt %d/%d failed: %v. Retrying in %v...\n",
					i+1, maxRetries, err, retryDelay,
				)
				timer := time.NewTimer(retryDelay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return nil, fmt.Errorf("database connection canceled: %w", ctx.Err())
				case <-timer.C:
				}
				if retryDelay < 8*time.Second {
					retryDelay *= 2
				}
			}
		}
	}

	return nil, fmt.Errorf(
		"failed to connect to database after %d attempts: %w",
		maxRetries,
		err,
	)
}

// configurePool sets up the database connection pool with values from configuration.
func configurePool(sqlDB *sql.DB, cfg config.DatabaseConfig) {
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	log.Printf(
		"Database pool configured: MaxIdleConns=%d, MaxOpenConns=%d, ConnMaxLifetime=%v",
		cfg.MaxIdleConns,
		cfg.MaxOpenConns,
		cfg.ConnMaxLifetime,
	)
}

// newGormLogger returns a GORM logger configured with the specified log level.
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

	log.Printf("GORM log level set to: %s", level)
	return glogger.Default.LogMode(lvl)
}
