package database

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/url"

	"payment-service/config"
	"payment-service/models"
)

func GetConnectionString(cfg *config.DatabaseConfig) (connStr string) {
	dsn := url.URL{
		Scheme: cfg.Driver,
		User:   url.UserPassword(cfg.Postgres.User, cfg.Postgres.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Postgres.Host, cfg.Postgres.Port),
		Path:   cfg.Postgres.Database,
	}

	q := dsn.Query()
	q.Set("sslmode", "disable")
	dsn.RawQuery = q.Encode()

	return dsn.String()
}

func Init(url string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})

	if err != nil {
		panic("database init")
	}

	db.AutoMigrate(&models.Payment{})

	return db
}
