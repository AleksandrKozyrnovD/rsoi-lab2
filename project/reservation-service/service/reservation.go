package service

import (
	"errors"
	"fmt"
	"time"

	"reservation-service/models"
	"reservation-service/repository"
)

var (
	errNotFound                   = errors.New("not found")
	ErrHotelNotFound              = errors.New("hotel not found")
	ErrReservationNotFound        = errors.New("reservation not found")
	ErrInvalidData                = errors.New("invalid data")
	ErrReservationAlreadyCanceled = errors.New("reservation already canceled")
)

type ReservationService interface {
	ListHotels(page, size int) (models.PaginationResponse, error)
	ListReservations(username string) ([]models.ReservationResponse, error)
	GetReservation(username, reservationUid string) (models.ReservationResponse, error)
	CreateReservation(username string, req models.CreateReservationRequest) (models.CreateReservationResponse, error)
	CancelReservation(username, reservationUid string) error
}

const dateLayout = "2006-01-02"

type reservationService struct {
	hotelRepo       repository.HotelRepository
	reservationRepo repository.ReservationRepository
}

// NewReservationService — конструктор сервиса.
func NewReservationService(
	hotelRepo repository.HotelRepository,
	reservationRepo repository.ReservationRepository,
) ReservationService {
	return &reservationService{
		hotelRepo:       hotelRepo,
		reservationRepo: reservationRepo,
	}
}

func (s *reservationService) ListHotels(page, size int) (models.PaginationResponse, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	hotels, total, err := s.hotelRepo.List(page, size)
	if err != nil {
		return models.PaginationResponse{}, err
	}

	items := make([]models.HotelResponse, 0, len(hotels))
	for _, h := range hotels {
		items = append(items, models.HotelResponse{
			HotelUid: h.HotelUid,
			Name:     h.Name,
			Country:  h.Country,
			City:     h.City,
			Address:  h.Address,
			Stars:    h.Stars,
			Price:    h.Price,
		})
	}

	return models.PaginationResponse{
		Page:          page,
		PageSize:      size,
		TotalElements: total,
		Items:         items,
	}, nil
}

func (s *reservationService) ListReservations(username string) ([]models.ReservationResponse, error) {
	reservations, err := s.reservationRepo.ListByUsername(username)
	if err != nil {
		return nil, err
	}

	result := make([]models.ReservationResponse, 0, len(reservations))
	for _, r := range reservations {
		result = append(result, toReservationResponse(r))
	}
	return result, nil
}

func (s *reservationService) GetReservation(username, reservationUid string) (models.ReservationResponse, error) {
	r, err := s.reservationRepo.GetByUsernameAndUID(username, reservationUid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.ReservationResponse{}, ErrReservationNotFound
		}
		return models.ReservationResponse{}, err
	}
	return toReservationResponse(*r), nil
}

func (s *reservationService) CreateReservation(
	username string,
	req models.CreateReservationRequest,
) (models.CreateReservationResponse, error) {
	if username == "" || req.HotelUid == "" || req.StartDate == "" || req.EndDate == "" {
		return models.CreateReservationResponse{}, ErrInvalidData
	}

	startDate, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		return models.CreateReservationResponse{}, ErrInvalidData
	}
	endDate, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		return models.CreateReservationResponse{}, ErrInvalidData
	}
	if !endDate.After(startDate) {
		return models.CreateReservationResponse{}, ErrInvalidData
	}

	// Отель всё ещё проверяем, чтобы не создавать бронь на несуществующий отель.
	hotel, err := s.hotelRepo.GetByUID(req.HotelUid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.CreateReservationResponse{}, ErrHotelNotFound
		}
		return models.CreateReservationResponse{}, err
	}

	// Значения по умолчанию — на случай, если gateway их не прислал.
	status := req.Status
	if status == "" {
		status = models.ReservationStatusPaid
	}
	paymentStatus := req.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = models.PaymentStatusPaid
	}

	reservation := &models.Reservation{
		Username:      username,
		HotelUid:      hotel.HotelUid,
		Hotel:         *hotel,
		StartDate:     startDate,
		EndDate:       endDate,
		Status:        status,
		PaymentStatus: paymentStatus,
		Price:         req.Price,
		Discount:      req.Discount,
	}

	created, err := s.reservationRepo.Create(reservation)
	if err != nil {
		return models.CreateReservationResponse{}, err
	}

	return models.CreateReservationResponse{
		ReservationUid: created.ReservationUid,
		HotelUid:       created.HotelUid,
		StartDate:      created.StartDate.Format(dateLayout),
		EndDate:        created.EndDate.Format(dateLayout),
		Discount:       created.Discount,
		Status:         created.Status,
		Payment: models.PaymentInfo{
			Status: created.PaymentStatus,
			Price:  created.Price,
		},
	}, nil
}

func (s *reservationService) CancelReservation(username, reservationUid string) error {
	r, err := s.reservationRepo.GetByUsernameAndUID(username, reservationUid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrReservationNotFound
		}
		return err
	}

	if r.Status == models.ReservationStatusCanceled {
		return ErrReservationAlreadyCanceled
	}

	if _, err := s.reservationRepo.Cancel(username, reservationUid); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrReservationNotFound
		}
		return err
	}
	return nil
}

func toReservationResponse(r models.Reservation) models.ReservationResponse {
	return models.ReservationResponse{
		ReservationUid: r.ReservationUid,
		Hotel: models.HotelInfo{
			HotelUid:    r.Hotel.HotelUid,
			Name:        r.Hotel.Name,
			FullAddress: fmt.Sprintf("%s, %s, %s", r.Hotel.Country, r.Hotel.City, r.Hotel.Address),
			Stars:       r.Hotel.Stars,
		},
		StartDate: r.StartDate.Format(dateLayout),
		EndDate:   r.EndDate.Format(dateLayout),
		Status:    r.Status,
		Payment: models.PaymentInfo{
			Status: r.PaymentStatus,
			Price:  r.Price,
		},
	}
}
