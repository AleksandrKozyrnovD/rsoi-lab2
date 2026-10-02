package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"loyalty-service/models"
)

var (
	ErrNotFound = errors.New("person not found")
)

type LoyaltyRepository interface {
	Create(loyalty *models.Loyalty) (*models.Loyalty, error)
	GetByUsername(username string) (*models.Loyalty, error)
	IncrementReservation(username string) (*models.Loyalty, error)
	DecrementReservation(username string) (*models.Loyalty, error)
}

type loyaltyRepository struct {
	DB *gorm.DB
}

func NewLoyaltyRepository(db *gorm.DB) LoyaltyRepository {
	return &loyaltyRepository{DB: db}
}

func (r *loyaltyRepository) Create(loyalty *models.Loyalty) (*models.Loyalty, error) {
	if loyalty == nil {
		return nil, errors.New("loyalty is nil")
	}

	if err := r.DB.Create(loyalty).Error; err != nil {
		return nil, err
	}

	return loyalty, nil
}

func (r *loyaltyRepository) GetByUsername(username string) (*models.Loyalty, error) {
	var loyalty models.Loyalty

	if err := r.DB.Where("username = ?", username).First(&loyalty).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("loyalty %s: %w", username, ErrNotFound)
		}
		return nil, fmt.Errorf("loyalty %s: %w", username, err)
	}

	return &loyalty, nil
}

func (r *loyaltyRepository) IncrementReservation(username string) (*models.Loyalty, error) {
	result := r.DB.Model(&models.Loyalty{}).
		Where("username = ?", username).
		UpdateColumn("reservation_count", gorm.Expr("reservation_count + ?", 1))

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("loyalty %s: %w", username, ErrNotFound)
	}

	return r.GetByUsername(username)
}

func (r *loyaltyRepository) DecrementReservation(username string) (*models.Loyalty, error) {
	result := r.DB.Model(&models.Loyalty{}).
		Where("username = ?", username).
		UpdateColumn("reservation_count", gorm.Expr("reservation_count - ?", 1))

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("loyalty %s: %w", username, ErrNotFound)
	}

	return r.GetByUsername(username)
}
