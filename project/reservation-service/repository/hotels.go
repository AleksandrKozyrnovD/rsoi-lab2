package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"reservation-service/models"
)

var (
	ErrNotFound = errors.New("not found")
)

type HotelRepository interface {
	List(page, size int) ([]models.Hotel, int64, error)
	GetByUID(hotelUid string) (*models.Hotel, error)
}

type hotelRepository struct {
	DB *gorm.DB
}

func NewHotelRepository(db *gorm.DB) HotelRepository {
	return &hotelRepository{DB: db}
}

func (r *hotelRepository) List(page, size int) ([]models.Hotel, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}

	var (
		hotels []models.Hotel
		total  int64
	)

	if err := r.DB.Model(&models.Hotel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count hotels: %w", err)
	}

	offset := (page - 1) * size
	if err := r.DB.
		Order("name ASC").
		Limit(size).
		Offset(offset).
		Find(&hotels).Error; err != nil {
		return nil, 0, fmt.Errorf("list hotels: %w", err)
	}

	return hotels, total, nil
}

func (r *hotelRepository) GetByUID(hotelUid string) (*models.Hotel, error) {
	var hotel models.Hotel

	if err := r.DB.Where("hotel_uid = ?", hotelUid).First(&hotel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("hotel %s: %w", hotelUid, ErrNotFound)
		}
		return nil, fmt.Errorf("hotel %s: %w", hotelUid, err)
	}

	return &hotel, nil
}
