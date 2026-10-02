package main

import (
	"context"
	"fmt"
	"gateway/api"
	"gateway/config"
	"gateway/logger"
	"gateway/middleware"
	"log"
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

	// Контроллер
	controller := api.NewV1(
		config.URLs.Loyalty,
		config.URLs.Reservation,
		config.URLs.Payment,
	)
	port := config.HTTP.Port

	r := gin.New()

	r.Use(middleware.SlogMiddleware())
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/hotels", controller.HandleGetHotels)

		v1.GET("/me", controller.HandleGetMe)

		v1.GET("/reservations", controller.HandleGetReservations)
		v1.POST("/reservations", controller.HandleCreateReservation)
		v1.GET("/reservations/:reservationUid", controller.HandleGetReservationByUid)
		v1.DELETE("/reservations/:reservationUid", controller.HandleCancelReservation)

		v1.GET("/loyalty", controller.HandleGetLoyalty)

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
