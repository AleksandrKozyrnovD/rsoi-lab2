package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"payment-service/api"
	"payment-service/config"
	"payment-service/database"
	"payment-service/logger"
	"payment-service/middleware"
	"payment-service/repository"
	"payment-service/service"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const configPath string = "./config.yml"

func main() {
	logger.Init() // pass settings and/or name/path

	config, err := config.InitConfig(configPath)
	if err != nil {
		log.Fatalf("init config: %v", err)
	}

	// Инициализация базы данных
	urlDatabase := database.GetConnectionString(&config.Database)
	db := database.Init(urlDatabase)

	// Репозиторий и сервис лояльности
	paymentRepository := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(paymentRepository)

	// Контроллер
	controller := api.NewV1(paymentService)
	port := config.HTTP.Port

	r := gin.New()

	r.Use(middleware.SlogMiddleware())
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		v1.POST("/payment", controller.HandlePaymentCreate)
		v1.GET("/payment/:paymentUid", controller.HandlePaymentGetByUID)
		v1.DELETE("/payment/:paymentUid", controller.HandlePaymentCancel)
	}

	log.Printf("Running at %s:%d", config.HTTP.Host, config.HTTP.Port)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.HTTP.Host, config.HTTP.Port),
		Handler: r,
	}

	go func() {
		log.Printf("Server running on port %d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}
