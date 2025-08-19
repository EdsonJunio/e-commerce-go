// cmd/api/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"


	"e-commerce-go/internal/product/repository"
	producthttp "e-commerce-go/internal/product/delivery/http"
	"e-commerce-go/internal/product/service"
	"e-commerce-go/internal/shared/database"
	"e-commerce-go/internal/shared/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

)

func main() {
	// Inicializa a conexão com o banco de dados
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco de dados: %v", err)
	}

	// Configuração do Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middlewares globais
	r.Use(gin.Recovery())
	r.Use(requestid.New())

	// Configuração do CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Middleware de tratamento de erros
	r.Use(middleware.ErrorHandler())

	// Rota de health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"db":     "connected",
		})
	})

	// Inicializa e registra as rotas dos produtos
	productRepo := repository.NewProductRepository(db)
	productSvc := service.NewProductService(productRepo)
	productHandler := producthttp.NewProductHandler(productSvc)
	productHandler.RegisterRoutes(r)

	// Configuração do servidor HTTP
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Inicia o servidor em uma goroutine
	go func() {
		log.Println("Servidor iniciado na porta 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar o servidor: %v", err)
		}
	}()

	// Aguarda sinal de interrupção para desligar o servidor graciosamente
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Desligando o servidor...")

	// Cria um contexto com timeout para o desligamento
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Tenta desligar o servidor graciosamente
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erro ao desligar o servidor: %v", err)
	}

	log.Println("Servidor desligado")
}
