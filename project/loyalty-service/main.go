package main

import (
	"context"
	"fmt"
	"log"
	"loyalty-service/api"
	"loyalty-service/config"
	"loyalty-service/database"
	"loyalty-service/logger"
	"loyalty-service/middleware"
	"loyalty-service/repository"
	"loyalty-service/service"
	"net/http"
	"os"
	"os/signal"
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
	loyaltyRepository := repository.NewLoyaltyRepository(db)
	loyaltyService := service.NewLoyaltyService(loyaltyRepository)

	// Контроллер
	controller := api.NewV1(loyaltyService)
	port := config.HTTP.Port

	r := gin.New()

	r.Use(middleware.SlogMiddleware())
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		v1.POST("/loyalty", controller.HandleLoyaltyCreate)
		v1.GET("/loyalty/:username", controller.HandleLoyaltyGetByUsername)
		v1.POST("/loyalty/:username/reservations", controller.HandleLoyaltyIncrementReservations)
		v1.DELETE("/loyalty/:username/reservations", controller.HandleLoyaltyDecrementReservations)
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
