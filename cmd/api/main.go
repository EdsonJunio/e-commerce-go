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

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "e-commerce-go/docs"

	categoryHTTP "e-commerce-go/internal/catalog/delivery/http"
	"e-commerce-go/internal/catalog/repository"
	"e-commerce-go/internal/catalog/service"
	identityHTTP "e-commerce-go/internal/identity/delivery/http"
	identityRepo "e-commerce-go/internal/identity/repository"
	identitySvc "e-commerce-go/internal/identity/service"
	"e-commerce-go/internal/shared/cache"
	"e-commerce-go/internal/shared/config"
	"e-commerce-go/internal/shared/database"
	"e-commerce-go/internal/shared/middleware"
	sharedService "e-commerce-go/internal/shared/service"
	"e-commerce-go/internal/shared/transport"
	"e-commerce-go/pkg/logger"
)

func main() {
	cfg := config.Load()

	if err := logger.Init(logger.Config{
		Environment: cfg.Environment,
		Service:     cfg.AppName,
		Version:     cfg.Version,
	}); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	setGinMode(cfg.Environment)

	srv, db, rdb, err := buildServer(cfg)
	if err != nil {
		logger.L().Fatal("failed to build server", zap.Error(err))
	}

	runServer(srv, cfg)
	waitForShutdown(srv, db, rdb, cfg)
}

func setGinMode(env string) {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
}

func runServer(srv *http.Server, cfg *config.Config) {
	go func() {
		logger.L().Info("server started",
			zap.String("addr", srv.Addr),
			zap.String("env", cfg.Environment),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.L().Fatal("server crashed", zap.Error(err))
		}
	}()
}

func waitForShutdown(srv *http.Server, db *gorm.DB, rdb *cache.RedisClient, cfg *config.Config) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	logger.L().Info("shutdown signal received", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.L().Error("server forced to shutdown", zap.Error(err))
	}

	closeResources(db, rdb)
	logger.L().Info("server exited properly")
}

func closeResources(db *gorm.DB, rdb *cache.RedisClient) {
	// Close SQL DB
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.L().Error("failed to close database connection", zap.Error(err))
		} else {
			logger.L().Info("database connection closed")
		}
	}

	// Close Redis
	if err := rdb.Close(); err != nil {
		logger.L().Error("failed to close redis connection", zap.Error(err))
	} else {
		logger.L().Info("redis connection closed")
	}
}

// buildServer orchestrates the dependency injection and router setup
func buildServer(cfg *config.Config) (*http.Server, *gorm.DB, *cache.RedisClient, error) {
	// 1. Infrastructure
	db, err := database.ConnectDB()
	if err != nil {
		return nil, nil, nil, err
	}

	rdb, err := cache.NewRedisClient(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	// 2. Router & Middlewares
	r := gin.New()
	setupMiddlewares(r, cfg)

	// 3. Shared Services
	jwtService := sharedService.NewJWTService(cfg)
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// 4. Modules Setup
	setupIdentityModule(r, db, jwtService)
	setupCatalogModule(r, db, rdb, authMiddleware)

	// 5. Documentation & Health
	setupOpsRoutes(r, db, cfg)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return srv, db, rdb, nil
}

func setupMiddlewares(r *gin.Engine, cfg *config.Config) {
	// Prometheus
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	// Standard
	r.Use(requestid.New())
	r.Use(logger.RecoveryWithLogger())
	r.Use(logger.GinLoggerMiddleware())
	r.Use(middleware.ErrorHandler())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))
}

func setupOpsRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Swagger & Redoc
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", transport.RedocHandler)

	// Health Checks
	registerHealthEndpoints(r, db, cfg.Version)

	// Pprof
	if cfg.Environment != "production" || os.Getenv("ENABLE_PPROF") == "true" {
		pprof.Register(r, "/debug/pprof")
	}
}

func setupIdentityModule(r *gin.Engine, db *gorm.DB, jwtSvc sharedService.JWTService) {
	repo := identityRepo.NewUserRepository(db)
	svc := identitySvc.NewAuthService(repo, jwtSvc)
	handler := identityHTTP.NewAuthHandler(svc)
	handler.RegisterRoutes(r)
}

func setupCatalogModule(r *gin.Engine, db *gorm.DB, rdb *cache.RedisClient, auth *middleware.AuthMiddleware) {
	catRepo := repository.NewCategoryRepository(db, rdb)
	prodRepo := repository.NewProductRepository(db)
	prodSKURepo := repository.NewProductSkuRepository(db)

	catSvc := service.NewCategoryService(catRepo)
	prodSvc := service.NewProductService(prodRepo, catRepo)
	prodSKUSvc := service.NewProductSkuService(prodSKURepo)

	catHandler := categoryHTTP.NewCategoryHandler(catSvc)
	prodHandler := categoryHTTP.NewProductHandler(prodSvc)
	prodSKUHandler := categoryHTTP.NewProductSkuHandler(prodSKUSvc)

	catHandler.RegisterCategoryRoutes(r, auth)
	prodHandler.RegisterProductRoutes(r, auth)
	prodSKUHandler.RegisterProductSkuRoutes(r, auth)

}

// registerHealthEndpoints remains the same...
func registerHealthEndpoints(r *gin.Engine, db *gorm.DB, version string) {
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive", "version": version})
	})

	r.GET("/ready", func(c *gin.Context) {
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
