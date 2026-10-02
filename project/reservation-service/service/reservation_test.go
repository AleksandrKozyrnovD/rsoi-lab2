package service

import (
	"errors"
	"testing"
	"time"

	"reservation-service/models"
	"reservation-service/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------- Mocks ----------------------

type MockHotelRepository struct {
	mock.Mock
}

func (m *MockHotelRepository) List(page, size int) ([]models.Hotel, int64, error) {
	args := m.Called(page, size)
	var hotels []models.Hotel
	if v := args.Get(0); v != nil {
		hotels = v.([]models.Hotel)
	}
	return hotels, args.Get(1).(int64), args.Error(2)
}

func (m *MockHotelRepository) GetByUID(hotelUid string) (*models.Hotel, error) {
	args := m.Called(hotelUid)
	var hotel *models.Hotel
	if v := args.Get(0); v != nil {
		hotel = v.(*models.Hotel)
	}
	return hotel, args.Error(1)
}

type MockReservationRepository struct {
	mock.Mock
}

func (m *MockReservationRepository) Create(reservation *models.Reservation) (*models.Reservation, error) {
	args := m.Called(reservation)
	var r *models.Reservation
	if v := args.Get(0); v != nil {
		r = v.(*models.Reservation)
	}
	return r, args.Error(1)
}

func (m *MockReservationRepository) ListByUsername(username string) ([]models.Reservation, error) {
	args := m.Called(username)
	var reservations []models.Reservation
	if v := args.Get(0); v != nil {
		reservations = v.([]models.Reservation)
	}
	return reservations, args.Error(1)
}

func (m *MockReservationRepository) GetByUsernameAndUID(username, reservationUid string) (*models.Reservation, error) {
	args := m.Called(username, reservationUid)
	var r *models.Reservation
	if v := args.Get(0); v != nil {
		r = v.(*models.Reservation)
	}
	return r, args.Error(1)
}

func (m *MockReservationRepository) Cancel(username, reservationUid string) (*models.Reservation, error) {
	args := m.Called(username, reservationUid)
	var r *models.Reservation
	if v := args.Get(0); v != nil {
		r = v.(*models.Reservation)
	}
	return r, args.Error(1)
}

// ---------------------- Helpers ----------------------

func newService() (ReservationService, *MockHotelRepository, *MockReservationRepository) {
	hotelRepo := new(MockHotelRepository)
	reservationRepo := new(MockReservationRepository)
	svc := NewReservationService(hotelRepo, reservationRepo)
	return svc, hotelRepo, reservationRepo
}

func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse(dateLayout, s)
	require.NoError(t, err)
	return d
}

func sampleHotel() models.Hotel {
	return models.Hotel{
		HotelUid: "hotel-1",
		Name:     "Grand Hotel",
		Country:  "Russia",
		City:     "Moscow",
		Address:  "Tverskaya 1",
		Stars:    5,
		Price:    1000,
	}
}

func sampleReservation(t *testing.T) models.Reservation {
	t.Helper()
	return models.Reservation{
		ReservationUid: "res-1",
		Username:       "user-1",
		HotelUid:       "hotel-1",
		Hotel:          sampleHotel(),
		StartDate:      mustParseDate(t, "2024-06-01"),
		EndDate:        mustParseDate(t, "2024-06-10"),
		Status:         models.ReservationStatusPaid,
		PaymentStatus:  models.PaymentStatusPaid,
		Price:          9000,
		Discount:       10,
	}
}

// ---------------------- ListHotels ----------------------

func TestListHotels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, hotelRepo, _ := newService()

		hotels := []models.Hotel{sampleHotel()}
		hotelRepo.On("List", 1, 10).Return(hotels, int64(1), nil)

		resp, err := svc.ListHotels(1, 10)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
		assert.Equal(t, int64(1), resp.TotalElements)
		require.Len(t, resp.Items, 1)
		assert.Equal(t, "hotel-1", resp.Items[0].HotelUid)
		assert.Equal(t, "Grand Hotel", resp.Items[0].Name)
		assert.Equal(t, "Russia", resp.Items[0].Country)
		assert.Equal(t, 5, resp.Items[0].Stars)
		assert.Equal(t, 1000, resp.Items[0].Price)

		hotelRepo.AssertExpectations(t)
	})

	t.Run("defaults when page/size invalid", func(t *testing.T) {
		svc, hotelRepo, _ := newService()

		hotelRepo.On("List", 1, 10).Return([]models.Hotel{}, int64(0), nil)

		resp, err := svc.ListHotels(0, -5)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 10, resp.PageSize)
		assert.Empty(t, resp.Items)
		hotelRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		svc, hotelRepo, _ := newService()

		repoErr := errors.New("db error")
		hotelRepo.On("List", 2, 5).Return(nil, int64(0), repoErr)

		_, err := svc.ListHotels(2, 5)

		assert.ErrorIs(t, err, repoErr)
		hotelRepo.AssertExpectations(t)
	})
}

// ---------------------- ListReservations ----------------------

func TestListReservations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		reservations := []models.Reservation{sampleReservation(t)}
		reservationRepo.On("ListByUsername", "user-1").Return(reservations, nil)

		resp, err := svc.ListReservations("user-1")

		require.NoError(t, err)
		require.Len(t, resp, 1)
		assert.Equal(t, "res-1", resp[0].ReservationUid)
		assert.Equal(t, "hotel-1", resp[0].Hotel.HotelUid)
		assert.Equal(t, "Russia, Moscow, Tverskaya 1", resp[0].Hotel.FullAddress)
		assert.Equal(t, "2024-06-01", resp[0].StartDate)
		assert.Equal(t, "2024-06-10", resp[0].EndDate)
		assert.Equal(t, models.ReservationStatusPaid, resp[0].Status)
		assert.Equal(t, models.PaymentStatusPaid, resp[0].Payment.Status)
		assert.Equal(t, 9000, resp[0].Payment.Price)

		reservationRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		reservationRepo.On("ListByUsername", "user-x").Return([]models.Reservation{}, nil)

		resp, err := svc.ListReservations("user-x")

		require.NoError(t, err)
		assert.Empty(t, resp)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		repoErr := errors.New("db error")
		reservationRepo.On("ListByUsername", "user-1").Return(nil, repoErr)

		_, err := svc.ListReservations("user-1")

		assert.ErrorIs(t, err, repoErr)
		reservationRepo.AssertExpectations(t)
	})
}

// ---------------------- GetReservation ----------------------

func TestGetReservation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		r := sampleReservation(t)
		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "res-1").
			Return(&r, nil)

		resp, err := svc.GetReservation("user-1", "res-1")

		require.NoError(t, err)
		assert.Equal(t, "res-1", resp.ReservationUid)
		assert.Equal(t, "Grand Hotel", resp.Hotel.Name)
		assert.Equal(t, models.ReservationStatusPaid, resp.Status)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "unknown").
			Return(nil, repository.ErrNotFound)

		_, err := svc.GetReservation("user-1", "unknown")

		assert.ErrorIs(t, err, ErrReservationNotFound)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		repoErr := errors.New("db error")
		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "res-1").
			Return(nil, repoErr)

		_, err := svc.GetReservation("user-1", "res-1")

		assert.ErrorIs(t, err, repoErr)
		reservationRepo.AssertExpectations(t)
	})
}

// ---------------------- CreateReservation ----------------------

func TestCreateReservation(t *testing.T) {
	validReq := models.CreateReservationRequest{
		HotelUid:      "hotel-1",
		StartDate:     "2024-06-01",
		EndDate:       "2024-06-10",
		Discount:      10,
		Price:         9000,
		Status:        models.ReservationStatusPaid,
		PaymentStatus: models.PaymentStatusPaid,
	}

	t.Run("success", func(t *testing.T) {
		svc, hotelRepo, reservationRepo := newService()

		h := sampleHotel()
		hotelRepo.On("GetByUID", "hotel-1").Return(&h, nil)

		// Созданная бронь, которую вернёт репозиторий.
		created := sampleReservation(t)
		created.Status = validReq.Status
		created.PaymentStatus = validReq.PaymentStatus
		created.Price = validReq.Price
		created.Discount = validReq.Discount

		reservationRepo.
			On("Create", mock.MatchedBy(func(r *models.Reservation) bool {
				return r.Username == "user-1" &&
					r.HotelUid == "hotel-1" &&
					r.Status == models.ReservationStatusPaid &&
					r.PaymentStatus == models.PaymentStatusPaid &&
					r.Price == 9000 &&
					r.Discount == 10 &&
					r.StartDate.Equal(mustParseDate(t, "2024-06-01")) &&
					r.EndDate.Equal(mustParseDate(t, "2024-06-10"))
			})).
			Return(&created, nil)

		resp, err := svc.CreateReservation("user-1", validReq)

		require.NoError(t, err)
		assert.Equal(t, "res-1", resp.ReservationUid)
		assert.Equal(t, "hotel-1", resp.HotelUid)
		assert.Equal(t, "2024-06-01", resp.StartDate)
		assert.Equal(t, "2024-06-10", resp.EndDate)
		assert.Equal(t, 10, resp.Discount)
		assert.Equal(t, models.ReservationStatusPaid, resp.Status)
		assert.Equal(t, models.PaymentStatusPaid, resp.Payment.Status)
		assert.Equal(t, 9000, resp.Payment.Price)

		hotelRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("applies default statuses", func(t *testing.T) {
		svc, hotelRepo, reservationRepo := newService()

		req := validReq
		req.Status = ""
		req.PaymentStatus = ""

		h := sampleHotel()
		hotelRepo.On("GetByUID", "hotel-1").Return(&h, nil)

		created := sampleReservation(t)
		created.Status = models.ReservationStatusPaid
		created.PaymentStatus = models.PaymentStatusPaid

		reservationRepo.
			On("Create", mock.MatchedBy(func(r *models.Reservation) bool {
				return r.Status == models.ReservationStatusPaid &&
					r.PaymentStatus == models.PaymentStatusPaid
			})).
			Return(&created, nil)

		_, err := svc.CreateReservation("user-1", req)

		require.NoError(t, err)
		hotelRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("invalid data - empty fields", func(t *testing.T) {
		svc, _, _ := newService()

		cases := []models.CreateReservationRequest{
			{}, // всё пустое
			{HotelUid: "hotel-1"},
			{HotelUid: "hotel-1", StartDate: "2024-06-01"},
			{HotelUid: "hotel-1", StartDate: "2024-06-01", EndDate: "2024-06-10"}, // username empty проверяется в аргументе
		}

		for i, req := range cases {
			username := "user-1"
			if i == len(cases)-1 {
				username = "" // проверяем пустой username
			}
			_, err := svc.CreateReservation(username, req)
			assert.ErrorIs(t, err, ErrInvalidData, "case %d", i)
		}
	})

	t.Run("invalid data - bad date format", func(t *testing.T) {
		svc, _, _ := newService()

		req := validReq
		req.StartDate = "01-06-2024"

		_, err := svc.CreateReservation("user-1", req)

		assert.ErrorIs(t, err, ErrInvalidData)
	})

	t.Run("invalid data - end before start", func(t *testing.T) {
		svc, _, _ := newService()

		req := validReq
		req.StartDate = "2024-06-10"
		req.EndDate = "2024-06-01"

		_, err := svc.CreateReservation("user-1", req)

		assert.ErrorIs(t, err, ErrInvalidData)
	})

	t.Run("invalid data - end equals start", func(t *testing.T) {
		svc, _, _ := newService()

		req := validReq
		req.StartDate = "2024-06-10"
		req.EndDate = "2024-06-10"

		_, err := svc.CreateReservation("user-1", req)

		assert.ErrorIs(t, err, ErrInvalidData)
	})

	t.Run("hotel not found", func(t *testing.T) {
		svc, hotelRepo, _ := newService()

		hotelRepo.On("GetByUID", "hotel-1").Return(nil, repository.ErrNotFound)

		_, err := svc.CreateReservation("user-1", validReq)

		assert.ErrorIs(t, err, ErrHotelNotFound)
		hotelRepo.AssertExpectations(t)
	})

	t.Run("hotel repo error", func(t *testing.T) {
		svc, hotelRepo, _ := newService()

		repoErr := errors.New("db error")
		hotelRepo.On("GetByUID", "hotel-1").Return(nil, repoErr)

		_, err := svc.CreateReservation("user-1", validReq)

		assert.ErrorIs(t, err, repoErr)
		hotelRepo.AssertExpectations(t)
	})

	t.Run("reservation repo error", func(t *testing.T) {
		svc, hotelRepo, reservationRepo := newService()

		h := sampleHotel()
		hotelRepo.On("GetByUID", "hotel-1").Return(&h, nil)

		repoErr := errors.New("db error")
		reservationRepo.On("Create", mock.Anything).Return(nil, repoErr)

		_, err := svc.CreateReservation("user-1", validReq)

		assert.ErrorIs(t, err, repoErr)
		hotelRepo.AssertExpectations(t)
		reservationRepo.AssertExpectations(t)
	})
}

// ---------------------- CancelReservation ----------------------

func TestCancelReservation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		r := sampleReservation(t)
		r.Status = models.ReservationStatusPaid

		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "res-1").
			Return(&r, nil)

		canceled := r
		canceled.Status = models.ReservationStatusCanceled
		reservationRepo.
			On("Cancel", "user-1", "res-1").
			Return(&canceled, nil)

		err := svc.CancelReservation("user-1", "res-1")

		require.NoError(t, err)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "unknown").
			Return(nil, repository.ErrNotFound)

		err := svc.CancelReservation("user-1", "unknown")

		assert.ErrorIs(t, err, ErrReservationNotFound)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("already canceled", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		r := sampleReservation(t)
		r.Status = models.ReservationStatusCanceled

		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "res-1").
			Return(&r, nil)

		err := svc.CancelReservation("user-1", "res-1")

		assert.ErrorIs(t, err, ErrReservationAlreadyCanceled)
		// Cancel НЕ должен вызываться
		reservationRepo.AssertNotCalled(t, "Cancel", mock.Anything, mock.Anything)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("get repo error", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		repoErr := errors.New("db error")
		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "res-1").
			Return(nil, repoErr)

		err := svc.CancelReservation("user-1", "res-1")

		assert.ErrorIs(t, err, repoErr)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("cancel repo error", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		r := sampleReservation(t)
		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "res-1").
			Return(&r, nil)

		repoErr := errors.New("db error")
		reservationRepo.
			On("Cancel", "user-1", "res-1").
			Return(nil, repoErr)

		err := svc.CancelReservation("user-1", "res-1")

		assert.ErrorIs(t, err, repoErr)
		reservationRepo.AssertExpectations(t)
	})

	t.Run("cancel repo returns not found", func(t *testing.T) {
		svc, _, reservationRepo := newService()

		r := sampleReservation(t)
		reservationRepo.
			On("GetByUsernameAndUID", "user-1", "res-1").
			Return(&r, nil)

		reservationRepo.
			On("Cancel", "user-1", "res-1").
			Return(nil, repository.ErrNotFound)

		err := svc.CancelReservation("user-1", "res-1")

		assert.ErrorIs(t, err, ErrReservationNotFound)
		reservationRepo.AssertExpectations(t)
	})
}
