package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"payment-service/models"
	"payment-service/service"
)

type V1 struct {
	paymentService service.PaymentService
}

func NewV1(paymentService service.PaymentService) *V1 {
	return &V1{paymentService: paymentService}
}

// POST /api/v1/payment
func (a *V1) HandlePaymentCreate(c *gin.Context) {
	var req models.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Message: "Invalid data",
			Errors: []models.ErrorDescription{
				{Field: "body", Error: err.Error()},
			},
		})
		return
	}

	if errs := validatePaymentRequest(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Message: "Invalid data",
			Errors:  errs,
		})
		return
	}

	created, err := a.paymentService.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.Header("Location", "/api/v1/payment/"+created.PaymentUID.String())
	c.JSON(http.StatusCreated, toPaymentResponse(created))
}

// GET /api/v1/payment/{paymentUid}
func (a *V1) HandlePaymentGetByUID(c *gin.Context) {
	paymentUID := c.Param("paymentUid")
	if paymentUID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "paymentUid is required"})
		return
	}

	payment, err := a.paymentService.GetByUID(paymentUID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, toPaymentResponse(payment))
}

// DELETE /api/v1/payment/{paymentUid}
func (a *V1) HandlePaymentCancel(c *gin.Context) {
	paymentUID := c.Param("paymentUid")
	if paymentUID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "paymentUid is required"})
		return
	}

	_, err := a.paymentService.Cancel(paymentUID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ---------- helpers ----------

func validatePaymentRequest(req models.PaymentRequest) []models.ErrorDescription {
	errs := []models.ErrorDescription{}
	if req.Price <= 0 {
		errs = append(errs, models.ErrorDescription{
			Field: "price",
			Error: "price must be greater than 0",
		})
	}
	return errs
}

func toPaymentResponse(p models.Payment) models.PaymentResponse {
	return models.PaymentResponse{
		PaymentUID: p.PaymentUID.String(),
		Status:     p.Status,
		Price:      p.Price,
	}
}
