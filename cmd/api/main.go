package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	productHTTP "e-commerce-go/internal/product/delivery/http"
	"e-commerce-go/internal/product/repository"
	"e-commerce-go/internal/product/service"
	"e-commerce-go/internal/shared/config"
	"e-commerce-go/internal/shared/database"
	"e-commerce-go/internal/shared/middleware"
	"e-commerce-go/pkg/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Carrega as configurações
	cfg := config.Load()

	if err := logger.Init(logger.Config{
		Environment: cfg.Environment,
		Service:     cfg.AppName,
		Version:     cfg.Version,
	}); err != nil {
		log.Fatalf("Erro ao inicializar o logger: %v", err)
	}
	defer logger.Sync()

	// Configura o logger do Gin
	if os.Getenv("GIN_MODE") != "release" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializa a conexão com o banco de dados
	db, err := database.ConnectDB()
	if err != nil {
		logger.L().Fatal("Falha ao conectar ao banco de dados", zap.Error(err))
	}

	// Cria uma nova instância do Gin
	r := gin.New()

	// Middlewares globais
	r.Use(gin.Recovery())
	r.Use(requestid.New())

	r.Use(logger.GinLoggerMiddleware())

	r.Use(logger.RecoveryWithLogger())

	// Configuração do CORS
	corsCfg := cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}
	r.Use(cors.New(corsCfg))

	// Middleware de tratamento de erros
	r.Use(middleware.ErrorHandler())

	// Rota de health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// Inicializa e registra as rotas dos produtos
	productRepo := repository.NewProductRepository(db)
	productSvc := service.NewProductService(productRepo)
	productHandler := productHTTP.NewProductHandler(productSvc)
	productHandler.RegisterRoutes(r)

	// Configuração do servidor HTTP
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Canal para receber erros do servidor
	serverErrors := make(chan error, 1)

	// Inicia o servidor em uma goroutine
	go func() {
		logger.L().Info("Servidor iniciado", zap.String("port", cfg.Server.Port))
		serverErrors <- srv.ListenAndServe()
	}()

	// Canal para capturar sinais do sistema operacional
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// Aguarda por um sinal de erro ou de desligamento
	select {
	case err := <-serverErrors:
		logger.L().Fatal("Erro ao iniciar o servidor", zap.Error(err))

	case sig := <-osSignals:
		logger.L().Info("Sinal recebido, iniciando desligamento gracioso...",
			zap.String("signal", sig.String()),
		)

		// Cria um contexto com timeout para o desligamento
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		// Tenta desligar o servidor graciosamente
		if err := srv.Shutdown(ctx); err != nil {
			logger.L().Fatal("Erro ao desligar o servidor", zap.Error(err))
		}

		logger.L().Info("Servidor desligado com sucesso")
	}
}
