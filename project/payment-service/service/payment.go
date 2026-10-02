package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"payment-service/models"
	"payment-service/repository"
)

var (
	ErrNotFound = errors.New("payment not found")
)

type PaymentService interface {
	Create(req models.PaymentRequest) (models.Payment, error)
	GetByUID(paymentUID string) (models.Payment, error)
	Cancel(paymentUID string) (models.Payment, error)
}

type paymentService struct {
	repo repository.PaymentRepository
}

func NewPaymentService(repo repository.PaymentRepository) PaymentService {
	return &paymentService{repo: repo}
}

func (s *paymentService) Create(req models.PaymentRequest) (models.Payment, error) {
	if req.Price <= 0 {
		return models.Payment{}, fmt.Errorf("price must be positive")
	}

	payment := &models.Payment{
		PaymentUID: uuid.New(),
		Status:     models.PaymentStatusPaid,
		Price:      req.Price,
	}

	created, err := s.repo.Create(payment)
	if err != nil {
		return models.Payment{}, err
	}
	if created == nil {
		return models.Payment{}, fmt.Errorf("repository returned nil payment")
	}

	return *created, nil
}

func (s *paymentService) GetByUID(paymentUID string) (models.Payment, error) {
	uid, err := uuid.Parse(paymentUID)
	if err != nil {
		return models.Payment{}, fmt.Errorf("invalid payment uid %q: %w", paymentUID, err)
	}

	payment, err := s.repo.GetByUID(uid)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, ErrNotFound
		}
		return models.Payment{}, err
	}
	if payment == nil {
		return models.Payment{}, ErrNotFound
	}

	return *payment, nil
}

func (s *paymentService) Cancel(paymentUID string) (models.Payment, error) {
	uid, err := uuid.Parse(paymentUID)
	if err != nil {
		return models.Payment{}, fmt.Errorf("invalid payment uid %q: %w", paymentUID, err)
	}

	payment, err := s.repo.Cancel(uid)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, ErrNotFound
		}
		return models.Payment{}, err
	}
	if payment == nil {
		return models.Payment{}, ErrNotFound
	}

	return *payment, nil
}
