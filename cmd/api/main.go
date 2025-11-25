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

	categoryHTTP "e-commerce-go/internal/catalog/delivery/http"
	"e-commerce-go/internal/catalog/repository"
	"e-commerce-go/internal/catalog/service"
	"e-commerce-go/internal/shared/cache"
	"e-commerce-go/internal/shared/config"
	"e-commerce-go/internal/shared/database"
	"e-commerce-go/internal/shared/middleware"
	"e-commerce-go/internal/shared/transport"
	"e-commerce-go/pkg/logger"

	_ "e-commerce-go/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// Load Configuration (Single Source of Truth)
	cfg := config.Load()

	// Initialize Logger (Immediately to catch boot errors)
	if err := logger.Init(logger.Config{
		Environment: cfg.Environment,
		Service:     cfg.AppName,
		Version:     cfg.Version,
	}); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Global Gin Configuration
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Boot Application
	srv, db, rdb, err := buildServer(cfg)
	if err != nil {
		logger.L().Fatal("failed to build server", zap.Error(err))
	}

	// Start Server in Background
	serverErrors := make(chan error, 1)
	go func() {
		logger.L().Info("server started",
			zap.String("addr", srv.Addr),
			zap.String("env", cfg.Environment),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Await Shutdown Signal (Graceful Shutdown)
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.L().Fatal("server crashed", zap.Error(err))

	case sig := <-osSignals:
		logger.L().Info("shutdown signal received", zap.String("signal", sig.String()))

		// Context with Timeout to force shutdown if it takes too long
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Close HTTP connections (stop receiving new requests)
		if err := srv.Shutdown(ctx); err != nil {
			logger.L().Error("server forced to shutdown", zap.Error(err))
		}

		// Close Database connection (Cleanup resources)
		closeDBConnection(db)

		//CLOSE REDIS TOO (Graceful Shutdown)
		if err := rdb.Close(); err != nil {
			logger.L().Error("failed to close redis connection", zap.Error(err))
		} else {
			logger.L().Info("redis connection closed")
		}

		logger.L().Info("server exited properly")
	}
}

// buildServer configures dependencies, routes, and returns the ready-to-run server.
func buildServer(cfg *config.Config) (*http.Server, *gorm.DB, *cache.RedisClient, error) {
	// Database Connection
	db, err := database.ConnectDB()
	if err != nil {
		return nil, nil, nil, err
	}

	rdb, err := cache.NewRedisClient(cfg)
	if err != nil {
		// If Redis fails, the application does not go up.
		return nil, nil, nil, err
	}

	// Router Initialization
	r := gin.New()

	// Global Middlewares
	r.Use(requestid.New())
	r.Use(logger.RecoveryWithLogger())
	r.Use(logger.GinLoggerMiddleware())

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	r.Use(middleware.ErrorHandler())

	// Infrastructure (Health & Diagnostics)
	registerHealthEndpoints(r, db, cfg.Version)
	registerDiagnostics(r, cfg.Environment)

	// 1. Swagger UI (Devs - Try it out)
	// Accessible: http://localhost:8081/swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 2. Redoc UI (Consumers - Beautiful Docs)
	// Accessible: http://localhost:8081/docs
	r.GET("/docs", transport.RedocHandler)

	// Module: Catalog
	// Since Product depends on Category, instantiate Category first
	categoryRepo := repository.NewCategoryRepository(db, rdb)
	productRepo := repository.NewProductRepository(db)

	categorySvc := service.NewCategoryService(categoryRepo)
	productSvc := service.NewProductService(productRepo, categoryRepo)

	categoryHandler := categoryHTTP.NewCategoryHandler(categorySvc)
	productHandler := categoryHTTP.NewProductHandler(productSvc)

	categoryHandler.RegisterCategoryRoutes(r)
	productHandler.RegisterProductRoutes(r)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return srv, db, rdb, nil
}

// closeDBConnection retrieves the generic SQL interface and closes it.
func closeDBConnection(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		logger.L().Error("failed to get sql.DB for closing", zap.Error(err))
		return
	}
	if err := sqlDB.Close(); err != nil {
		logger.L().Error("failed to close database connection", zap.Error(err))
	} else {
		logger.L().Info("database connection closed")
	}
}

// registerHealthEndpoints exposes /live and /ready endpoints for k8s probes.
func registerHealthEndpoints(r *gin.Engine, db *gorm.DB, version string) {
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive", "version": version})
	})

	r.GET("/ready", func(c *gin.Context) {
		// Short timeout to avoid blocking the Load Balancer
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down", "reason": "db_handle_error"})
			return
		}

		if err := sqlDB.PingContext(ctx); err != nil {
			logger.L().Warn("readiness check failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down", "reason": "db_unreachable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready", "version": version})
	})
}

// registerDiagnostics conditionally enables pprof.
func registerDiagnostics(r *gin.Engine, env string) {
	// Pprof should be protected or disabled in public production environments
	if env != "production" || os.Getenv("ENABLE_PPROF") == "true" {
		pprof.Register(r, "/debug/pprof")
	}
}
