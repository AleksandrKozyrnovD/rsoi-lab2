package service_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"payment-service/models"
	"payment-service/repository"
	"payment-service/service"
)

// MockPaymentRepository — мок для repository.PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(payment *models.Payment) (*models.Payment, error) {
	args := m.Called(payment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByUID(paymentUID uuid.UUID) (*models.Payment, error) {
	args := m.Called(paymentUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Cancel(paymentUID uuid.UUID) (*models.Payment, error) {
	args := m.Called(paymentUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func setupTest() (*MockPaymentRepository, service.PaymentService) {
	mockRepo := new(MockPaymentRepository)
	svc := service.NewPaymentService(mockRepo)
	return mockRepo, svc
}

// ---------- Create ----------

func TestPaymentService_Create_Success(t *testing.T) {
	mockRepo, svc := setupTest()

	req := models.PaymentRequest{Price: 100}
	expectedPayment := &models.Payment{
		PaymentUID: uuid.New(),
		Status:     models.PaymentStatusPaid,
		Price:      100,
	}

	mockRepo.On("Create", mock.MatchedBy(func(p *models.Payment) bool {
		return p.Price == 100 && p.Status == models.PaymentStatusPaid && p.PaymentUID != uuid.Nil
	})).Return(expectedPayment, nil)

	payment, err := svc.Create(req)

	assert.NoError(t, err)
	assert.Equal(t, *expectedPayment, payment)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_Create_InvalidPrice(t *testing.T) {
	mockRepo, svc := setupTest()

	req := models.PaymentRequest{Price: 0}
	_, err := svc.Create(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "price must be positive")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestPaymentService_Create_RepoError(t *testing.T) {
	mockRepo, svc := setupTest()

	req := models.PaymentRequest{Price: 100}
	repoErr := errors.New("db error")
	mockRepo.On("Create", mock.Anything).Return(nil, repoErr)

	_, err := svc.Create(req)

	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_Create_RepoReturnsNil(t *testing.T) {
	mockRepo, svc := setupTest()

	req := models.PaymentRequest{Price: 100}
	mockRepo.On("Create", mock.Anything).Return(nil, nil)

	_, err := svc.Create(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository returned nil payment")
	mockRepo.AssertExpectations(t)
}

// ---------- GetByUID ----------

func TestPaymentService_GetByUID_Success(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	expectedPayment := &models.Payment{
		PaymentUID: uid,
		Status:     models.PaymentStatusPaid,
		Price:      100,
	}

	mockRepo.On("GetByUID", uid).Return(expectedPayment, nil)

	payment, err := svc.GetByUID(uid.String())

	assert.NoError(t, err)
	assert.Equal(t, *expectedPayment, payment)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_GetByUID_InvalidUUID(t *testing.T) {
	mockRepo, svc := setupTest()

	_, err := svc.GetByUID("invalid-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payment uid")
	mockRepo.AssertNotCalled(t, "GetByUID")
}

func TestPaymentService_GetByUID_NotFound(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	mockRepo.On("GetByUID", uid).Return(nil, repository.ErrPaymentNotFound)

	_, err := svc.GetByUID(uid.String())

	assert.Error(t, err)
	assert.Equal(t, service.ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_GetByUID_RepoError(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	repoErr := errors.New("db error")
	mockRepo.On("GetByUID", uid).Return(nil, repoErr)

	_, err := svc.GetByUID(uid.String())

	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_GetByUID_RepoReturnsNil(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	mockRepo.On("GetByUID", uid).Return(nil, nil)

	_, err := svc.GetByUID(uid.String())

	assert.Error(t, err)
	assert.Equal(t, service.ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

// ---------- Cancel ----------

func TestPaymentService_Cancel_Success(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	expectedPayment := &models.Payment{
		PaymentUID: uid,
		Status:     models.PaymentStatusCanceled,
		Price:      100,
	}

	mockRepo.On("Cancel", uid).Return(expectedPayment, nil)

	payment, err := svc.Cancel(uid.String())

	assert.NoError(t, err)
	assert.Equal(t, *expectedPayment, payment)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_Cancel_InvalidUUID(t *testing.T) {
	mockRepo, svc := setupTest()

	_, err := svc.Cancel("invalid-uuid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payment uid")
	mockRepo.AssertNotCalled(t, "Cancel")
}

func TestPaymentService_Cancel_NotFound(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	mockRepo.On("Cancel", uid).Return(nil, repository.ErrPaymentNotFound)

	_, err := svc.Cancel(uid.String())

	assert.Error(t, err)
	assert.Equal(t, service.ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_Cancel_RepoError(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	repoErr := errors.New("db error")
	mockRepo.On("Cancel", uid).Return(nil, repoErr)

	_, err := svc.Cancel(uid.String())

	assert.Error(t, err)
	assert.Equal(t, repoErr, err)
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_Cancel_RepoReturnsNil(t *testing.T) {
	mockRepo, svc := setupTest()

	uid := uuid.New()
	mockRepo.On("Cancel", uid).Return(nil, nil)

	_, err := svc.Cancel(uid.String())

	assert.Error(t, err)
	assert.Equal(t, service.ErrNotFound, err)
	mockRepo.AssertExpectations(t)
}
