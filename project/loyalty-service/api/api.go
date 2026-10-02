package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"loyalty-service/models"
	"loyalty-service/service"
)

type V1 struct {
	loyaltyService service.LoyaltyService
}

func NewV1(loyaltyService service.LoyaltyService) *V1 {
	return &V1{loyaltyService: loyaltyService}
}

// POST /api/v1/loyalty
func (a *V1) HandleLoyaltyCreate(c *gin.Context) {
	var req models.LoyaltyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Message: "Invalid data",
			Errors: []models.ErrorDescription{
				{Field: "body", Error: err.Error()},
			},
		})
		return
	}

	if errs := validateLoyaltyRequest(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Message: "Invalid data",
			Errors:  errs,
		})
		return
	}

	created, err := a.loyaltyService.Create(toServiceLoyalty(req))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Message: err.Error(),
			Errors:  []models.ErrorDescription{},
		})
		return
	}

	c.Header("Location", "/api/v1/loyalty/"+created.Username)
	c.JSON(http.StatusCreated, created)
}

// GET /api/v1/loyalty/{username}
func (a *V1) HandleLoyaltyGetByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "username is required"})
		return
	}

	loyalty, err := a.loyaltyService.GetByUsername(username)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Loyalty info not found for username"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toLoyaltyInfoResponse(loyalty))
}

// POST /api/v1/loyalty/{username}/reservations
func (a *V1) HandleLoyaltyIncrementReservations(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "username is required"})
		return
	}

	loyalty, err := a.loyaltyService.IncrementReservation(username)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Loyalty info not found for username"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toLoyaltyInfoResponse(loyalty))
}

// DELETE /api/v1/loyalty/{username}/reservations
func (a *V1) HandleLoyaltyDecrementReservations(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "username is required"})
		return
	}

	loyalty, err := a.loyaltyService.DecrementReservation(username)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Loyalty info not found for username"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toLoyaltyInfoResponse(loyalty))
}

// ---------- helpers ----------

func validateLoyaltyRequest(req models.LoyaltyRequest) []models.ErrorDescription {
	errs := []models.ErrorDescription{}
	if req.Username == "" {
		errs = append(errs, models.ErrorDescription{
			Field: "username",
			Error: "username is required",
		})
	}
	return errs
}

func toServiceLoyalty(req models.LoyaltyRequest) models.Loyalty {
	return models.Loyalty{
		Username:         req.Username,
		ReservationCount: req.ReservationCount,
		Status:           req.Status,
		Discount:         req.Discount,
	}
}

func toLoyaltyInfoResponse(l models.Loyalty) models.LoyaltyInfoResponse {
	return models.LoyaltyInfoResponse{
		Status:           l.Status,
		Discount:         l.Discount,
		ReservationCount: l.ReservationCount,
	}
}
