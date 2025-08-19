package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"e-commerce-go/internal/shared/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// ConnectDB cria uma nova conexão com o banco de dados
func ConnectDB() (*gorm.DB, error) {
	host := config.GetEnvString("DB_HOST", "localhost")
	port := config.GetEnvString("DB_PORT", "5432")
	user := config.GetEnvString("DB_USER", "postgres")
	password := config.GetEnvString("DB_PASSWORD", "1234")
	dbname := config.GetEnvString("DB_NAME", "postgres")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	gormCfg := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 gormLoggerFromEnv(),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("falha ao obter instância do banco de dados: %w", err)
	}

	configurePool(sqlDB)

	// Valida a conexão com timeout curto
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("falha ao validar conexão com o banco de dados: %w", err)
	}

	return db, nil
}

func configurePool(sqlDB *sql.DB) {
	maxOpen := config.GetEnvInt("DB_MAX_OPEN_CONNS", 10)
	maxIdle := config.GetEnvInt("DB_MAX_IDLE_CONNS", 5)
	lifetime := config.GetEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute)

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(lifetime)
}

func gormLoggerFromEnv() glogger.Interface {
	levelStr := strings.ToLower(config.GetEnvString("DB_LOG_LEVEL", "warn"))
	var lvl glogger.LogLevel
	switch levelStr {
	case "silent":
		lvl = glogger.Silent
	case "error":
		lvl = glogger.Error
	case "info":
		lvl = glogger.Info
	default:
		lvl = glogger.Warn
	}
	return glogger.Default.LogMode(lvl)
}
