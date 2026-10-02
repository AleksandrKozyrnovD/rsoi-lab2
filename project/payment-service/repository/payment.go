package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"payment-service/models"
)

var (
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrPaymentAlreadyCanceled  = errors.New("payment already canceled")
	ErrPaymentCannotBeCanceled = errors.New("payment cannot be canceled")
)

type PaymentRepository interface {
	Create(payment *models.Payment) (*models.Payment, error)
	GetByUID(paymentUID uuid.UUID) (*models.Payment, error)
	Cancel(paymentUID uuid.UUID) (*models.Payment, error)
}

type paymentRepository struct {
	DB *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{DB: db}
}

func (r *paymentRepository) Create(payment *models.Payment) (*models.Payment, error) {
	if payment == nil {
		return nil, errors.New("payment is nil")
	}

	if err := r.DB.Create(payment).Error; err != nil {
		return nil, err
	}

	return payment, nil
}

func (r *paymentRepository) GetByUID(paymentUID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment

	if err := r.DB.Where("payment_uid = ?", paymentUID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("payment %s: %w", paymentUID, ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("payment %s: %w", paymentUID, err)
	}

	return &payment, nil
}

func (r *paymentRepository) Cancel(paymentUID uuid.UUID) (*models.Payment, error) {
	result := r.DB.Model(&models.Payment{}).
		Where("payment_uid = ? AND status = ?", paymentUID, models.PaymentStatusPaid).
		UpdateColumn("status", models.PaymentStatusCanceled)

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		existing, err := r.GetByUID(paymentUID)
		if err != nil {
			return nil, err
		}

		if existing.Status == models.PaymentStatusCanceled {
			return nil, fmt.Errorf("payment %s: %w", paymentUID, ErrPaymentAlreadyCanceled)
		}

		return nil, fmt.Errorf(
			"payment %s: status %s: %w",
			paymentUID,
			existing.Status,
			ErrPaymentCannotBeCanceled,
		)
	}

	return r.GetByUID(paymentUID)
}
