package service

import (
	"errors"
	"loyalty-service/models"
	"loyalty-service/repository"
)

var (
	ErrNotFound = errors.New("person not found")
)

type LoyaltyService interface {
	Create(models.Loyalty) (models.Loyalty, error)
	GetByUsername(username string) (models.Loyalty, error)
	IncrementReservation(username string) (models.Loyalty, error)
	DecrementReservation(username string) (models.Loyalty, error)
}

type loyaltyService struct {
	repo repository.LoyaltyRepository
}

func NewLoyaltyService(loyaltyRepo repository.LoyaltyRepository) LoyaltyService {
	return &loyaltyService{
		repo: loyaltyRepo,
	}
}

func (s *loyaltyService) Create(loyalty models.Loyalty) (models.Loyalty, error) {
	created, err := s.repo.Create(&loyalty)
	if err != nil {
		return models.Loyalty{}, err
	}
	if created == nil {
		return models.Loyalty{}, errors.New("created loyalty is nil")
	}

	return *created, nil
}

func (s *loyaltyService) GetByUsername(username string) (models.Loyalty, error) {
	loyalty, err := s.repo.GetByUsername(username)
	if err != nil {
		return models.Loyalty{}, mapRepositoryError(err)
	}
	if loyalty == nil {
		return models.Loyalty{}, ErrNotFound
	}

	return *loyalty, nil
}

func (s *loyaltyService) IncrementReservation(username string) (models.Loyalty, error) {
	loyalty, err := s.repo.IncrementReservation(username)
	if err != nil {
		return models.Loyalty{}, mapRepositoryError(err)
	}
	if loyalty == nil {
		return models.Loyalty{}, ErrNotFound
	}

	return *loyalty, nil
}

func (s *loyaltyService) DecrementReservation(username string) (models.Loyalty, error) {
	loyalty, err := s.repo.DecrementReservation(username)
	if err != nil {
		return models.Loyalty{}, mapRepositoryError(err)
	}
	if loyalty == nil {
		return models.Loyalty{}, ErrNotFound
	}

	return *loyalty, nil
}

func mapRepositoryError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
