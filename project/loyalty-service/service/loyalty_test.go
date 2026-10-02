package service_test

import (
	"errors"
	"testing"

	"loyalty-service/models"
	"loyalty-service/repository"
	"loyalty-service/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------- Mock ----------

type MockLoyaltyRepository struct {
	mock.Mock
}

func (m *MockLoyaltyRepository) Create(loyalty *models.Loyalty) (*models.Loyalty, error) {
	args := m.Called(loyalty)
	if v := args.Get(0); v != nil {
		return v.(*models.Loyalty), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLoyaltyRepository) GetByUsername(username string) (*models.Loyalty, error) {
	args := m.Called(username)
	if v := args.Get(0); v != nil {
		return v.(*models.Loyalty), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLoyaltyRepository) IncrementReservation(username string) (*models.Loyalty, error) {
	args := m.Called(username)
	if v := args.Get(0); v != nil {
		return v.(*models.Loyalty), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockLoyaltyRepository) DecrementReservation(username string) (*models.Loyalty, error) {
	args := m.Called(username)
	if v := args.Get(0); v != nil {
		return v.(*models.Loyalty), args.Error(1)
	}
	return nil, args.Error(1)
}

// ---------- Create ----------

func TestCreate_Success(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	input := models.Loyalty{Username: "user1"}
	expected := models.Loyalty{Username: "user1", ReservationCount: 5}

	repo.
		On("Create", mock.MatchedBy(func(l *models.Loyalty) bool {
			return l != nil && l.Username == "user1"
		})).
		Return(&expected, nil).
		Once()

	got, err := svc.Create(input)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestCreate_RepositoryError(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repoErr := errors.New("db error")
	repo.On("Create", mock.AnythingOfType("*models.Loyalty")).
		Return(nil, repoErr).
		Once()

	got, err := svc.Create(models.Loyalty{Username: "user1"})

	assert.ErrorIs(t, err, repoErr)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

func TestCreate_NilResultWithoutError(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repo.On("Create", mock.AnythingOfType("*models.Loyalty")).
		Return(nil, nil).
		Once()

	got, err := svc.Create(models.Loyalty{Username: "user1"})

	require.Error(t, err)
	assert.EqualError(t, err, "created loyalty is nil")
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

// ---------- GetByUsername ----------

func TestGetByUsername_Success(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	expected := models.Loyalty{Username: "user1", ReservationCount: 3}
	repo.On("GetByUsername", "user1").Return(&expected, nil).Once()

	got, err := svc.GetByUsername("user1")

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestGetByUsername_RepoErrNotFound_IsMapped(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repo.On("GetByUsername", "user1").
		Return(nil, repository.ErrNotFound).
		Once()

	got, err := svc.GetByUsername("user1")

	assert.ErrorIs(t, err, service.ErrNotFound)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

func TestGetByUsername_RepoOtherError_IsPassedThrough(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repoErr := errors.New("connection lost")
	repo.On("GetByUsername", "user1").Return(nil, repoErr).Once()

	got, err := svc.GetByUsername("user1")

	assert.ErrorIs(t, err, repoErr)
	assert.NotErrorIs(t, err, service.ErrNotFound)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

func TestGetByUsername_NilLoyaltyWithoutError(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repo.On("GetByUsername", "user1").Return(nil, nil).Once()

	got, err := svc.GetByUsername("user1")

	assert.ErrorIs(t, err, service.ErrNotFound)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

// ---------- IncrementReservation ----------

func TestIncrementReservation_Success(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	expected := models.Loyalty{Username: "user1", ReservationCount: 2}
	repo.On("IncrementReservation", "user1").Return(&expected, nil).Once()

	got, err := svc.IncrementReservation("user1")

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestIncrementReservation_RepoErrNotFound_IsMapped(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repo.On("IncrementReservation", "user1").
		Return(nil, repository.ErrNotFound).
		Once()

	got, err := svc.IncrementReservation("user1")

	assert.ErrorIs(t, err, service.ErrNotFound)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

func TestIncrementReservation_NilLoyaltyWithoutError(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repo.On("IncrementReservation", "user1").Return(nil, nil).Once()

	got, err := svc.IncrementReservation("user1")

	assert.ErrorIs(t, err, service.ErrNotFound)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

// ---------- DecrementReservation ----------

func TestDecrementReservation_Success(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	expected := models.Loyalty{Username: "user1", ReservationCount: 0}
	repo.On("DecrementReservation", "user1").Return(&expected, nil).Once()

	got, err := svc.DecrementReservation("user1")

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestDecrementReservation_RepoErrNotFound_IsMapped(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repo.On("DecrementReservation", "user1").
		Return(nil, repository.ErrNotFound).
		Once()

	got, err := svc.DecrementReservation("user1")

	assert.ErrorIs(t, err, service.ErrNotFound)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

func TestDecrementReservation_NilLoyaltyWithoutError(t *testing.T) {
	repo := new(MockLoyaltyRepository)
	svc := service.NewLoyaltyService(repo)

	repo.On("DecrementReservation", "user1").Return(nil, nil).Once()

	got, err := svc.DecrementReservation("user1")

	assert.ErrorIs(t, err, service.ErrNotFound)
	assert.Equal(t, models.Loyalty{}, got)
	repo.AssertExpectations(t)
}

// ---------- mapRepositoryError через разные методы ----------

func TestErrorMapping_ConsistentAcrossMethods(t *testing.T) {
	cases := []struct {
		name  string
		setup func(repo *MockLoyaltyRepository)
		call  func(svc service.LoyaltyService) error
	}{
		{
			name: "GetByUsername",
			setup: func(r *MockLoyaltyRepository) {
				r.On("GetByUsername", "u").Return(nil, repository.ErrNotFound).Once()
			},
			call: func(s service.LoyaltyService) error {
				_, err := s.GetByUsername("u")
				return err
			},
		},
		{
			name: "IncrementReservation",
			setup: func(r *MockLoyaltyRepository) {
				r.On("IncrementReservation", "u").Return(nil, repository.ErrNotFound).Once()
			},
			call: func(s service.LoyaltyService) error {
				_, err := s.IncrementReservation("u")
				return err
			},
		},
		{
			name: "DecrementReservation",
			setup: func(r *MockLoyaltyRepository) {
				r.On("DecrementReservation", "u").Return(nil, repository.ErrNotFound).Once()
			},
			call: func(s service.LoyaltyService) error {
				_, err := s.DecrementReservation("u")
				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockLoyaltyRepository)
			svc := service.NewLoyaltyService(repo)

			tc.setup(repo)
			err := tc.call(svc)

			assert.ErrorIs(t, err, service.ErrNotFound)
			repo.AssertExpectations(t)
		})
	}
}
