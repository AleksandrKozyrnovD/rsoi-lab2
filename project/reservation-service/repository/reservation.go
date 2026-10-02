package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"reservation-service/models"
)

type ReservationRepository interface {
	Create(reservation *models.Reservation) (*models.Reservation, error)
	ListByUsername(username string) ([]models.Reservation, error)
	GetByUsernameAndUID(username, reservationUid string) (*models.Reservation, error)
	Cancel(username, reservationUid string) (*models.Reservation, error)
}

type reservationRepository struct {
	DB *gorm.DB
}

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{DB: db}
}

func (r *reservationRepository) Create(reservation *models.Reservation) (*models.Reservation, error) {
	if reservation == nil {
		return nil, errors.New("reservation is nil")
	}

	if err := r.DB.Create(reservation).Error; err != nil {
		return nil, fmt.Errorf("create reservation: %w", err)
	}

	return reservation, nil
}

func (r *reservationRepository) ListByUsername(username string) ([]models.Reservation, error) {
	var reservations []models.Reservation

	if err := r.DB.
		Preload("Hotel").
		Where("username = ?", username).
		Order("start_date DESC").
		Find(&reservations).Error; err != nil {
		return nil, fmt.Errorf("list reservations for %s: %w", username, err)
	}

	return reservations, nil
}

func (r *reservationRepository) GetByUsernameAndUID(username, reservationUid string) (*models.Reservation, error) {
	var reservation models.Reservation

	if err := r.DB.
		Preload("Hotel").
		Where("username = ? AND reservation_uid = ?", username, reservationUid).
		First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("reservation %s: %w", reservationUid, ErrNotFound)
		}
		return nil, fmt.Errorf("reservation %s: %w", reservationUid, err)
	}

	return &reservation, nil
}

func (r *reservationRepository) Cancel(username, reservationUid string) (*models.Reservation, error) {
	result := r.DB.Model(&models.Reservation{}).
		Where("username = ? AND reservation_uid = ?", username, reservationUid).
		Updates(map[string]interface{}{
			"status":         models.ReservationStatusCanceled,
			"payment_status": models.PaymentStatusReversed,
		})

	if result.Error != nil {
		return nil, fmt.Errorf("cancel reservation %s: %w", reservationUid, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("reservation %s: %w", reservationUid, ErrNotFound)
	}

	return r.GetByUsernameAndUID(username, reservationUid)
}
