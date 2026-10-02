package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reservation-service/api"
	"reservation-service/config"
	"reservation-service/database"
	"reservation-service/logger"
	"reservation-service/middleware"
	"reservation-service/repository"
	"reservation-service/service"
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

	reservationRepository := repository.NewReservationRepository(db)
	hotelRepository := repository.NewHotelRepository(db)
	reservationService := service.NewReservationService(hotelRepository, reservationRepository)

	// Контроллер
	controller := api.NewV1(reservationService)
	port := config.HTTP.Port

	r := gin.New()

	r.Use(middleware.SlogMiddleware())
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/hotels", controller.HandleHotelsList)

		v1.GET("/reservations", controller.HandleReservationList)
		v1.POST("/reservations", controller.HandleReservationCreate)
		v1.GET("/reservations/:reservationUid", controller.HandleReservationGet)
		v1.DELETE("/reservations/:reservationUid", controller.HandleReservationCancel)

		r.GET("/manage/health", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
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
