package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	productHTTP "e-commerce-go/internal/product/delivery/http"
	"e-commerce-go/internal/product/repository"
	"e-commerce-go/internal/product/service"
	"e-commerce-go/internal/shared/config"
	"e-commerce-go/internal/shared/database"
	"e-commerce-go/internal/shared/middleware"
	"e-commerce-go/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// bootServer initializes dependencies, registers routes, and returns a configured *http.Server.
func bootServer() (*http.Server, *gorm.DB, error) {
	// Load configuration.
	cfg := config.Load()

	// Initialize global logger.
	if err := logger.Init(logger.Config{
		Environment: cfg.Environment,
		Service:     cfg.AppName,
		Version:     cfg.Version,
	}); err != nil {
		return nil, nil, err
	}

	// Configure Gin mode based on environment.
	if os.Getenv("GIN_MODE") != "release" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize DB connection.
	db, err := database.ConnectDB()
	if err != nil {
		logger.L().Fatal("failed to connect to database", zap.Error(err))
	}

	// Create Gin engine.
	r := gin.New()

	// Global middlewares.
	r.Use(requestid.New())
	r.Use(logger.RecoveryWithLogger())  // recovers + logs stack
	r.Use(logger.GinLoggerMiddleware()) // access logs

	// CORS.
	corsCfg := cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}
	r.Use(cors.New(corsCfg))

	// Error normalization.
	r.Use(middleware.ErrorHandler())

	// Health & diagnostics.
	registerHealthEndpoints(r, db, cfg.Version)
	registerDiagnostics(r, cfg.Environment)

	// Product module wiring.
	productRepo := repository.NewProductRepository(db)
	productSvc := service.NewProductService(productRepo)
	productHandler := productHTTP.NewProductHandler(productSvc)
	productHandler.RegisterRoutes(r)

	// Build HTTP server.
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return srv, db, nil
}

// registerHealthEndpoints exposes /live and /ready endpoints.
// - /live: always 200 if the process is running
// - /ready: pings the DB with a short timeout; 200 if ready, 503 otherwise
func registerHealthEndpoints(r *gin.Engine, db *gorm.DB, version string) {
	// Liveness: basic process health.
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "alive",
			"version": version,
		})
	})

	// Readiness: check critical deps (DB).
	r.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		sqlDB, err := db.DB()
		if err != nil {
			logger.L().Error("readiness: failed to get sqlDB", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not_ready",
				"version": version,
				"reason":  "db_handle_error",
			})
			return
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			logger.L().Warn("readiness: db ping failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not_ready",
				"version": version,
				"reason":  "db_unreachable",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"version": version,
		})
	})

	// Simple health (kept for compatibility).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": version,
		})
	})
}

// registerDiagnostics conditionally registers pprof under /debug/pprof.
// Enabled when ENV != production OR when ENABLE_PPROF=true.
func registerDiagnostics(r *gin.Engine, environment string) {
	if environment != "production" || os.Getenv("ENABLE_PPROF") == "true" {
		pprof.Register(r, "/debug/pprof")
		logger.L().Info("pprof enabled at /debug/pprof", zap.String("env", environment))
	}
}

func main() {
	srv, _, err := bootServer()
	if err != nil {
		log.Fatalf("failed to boot server: %v", err)
	}
	defer logger.Sync()

	// Start server in background.
	serverErrors := make(chan error, 1)
	go func() {
		logger.L().Info("server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}
		serverErrors <- nil
	}()

	// OS signal handling for graceful shutdown.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil {
			logger.L().Fatal("server failed to start", zap.Error(err))
		} else {
			logger.L().Info("server stopped")
		}

	case sig := <-osSignals:
		logger.L().Info("shutdown signal received, starting graceful shutdown",
			zap.String("signal", sig.String()),
		)

		// Graceful shutdown with timeout from config.
		cfg := config.Load()
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.L().Fatal("server shutdown error", zap.Error(err))
		}
		logger.L().Info("server shut down gracefully")
	}
}
