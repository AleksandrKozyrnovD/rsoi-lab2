package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"reservation-service/models"
	"reservation-service/service"
)

type V1 struct {
	reservationService service.ReservationService
}

func NewV1(reservationService service.ReservationService) *V1 {
	return &V1{reservationService: reservationService}
}

// GET /api/v1/hotels?page=&size=
func (a *V1) HandleHotelsList(c *gin.Context) {
	page := 0
	size := 20

	if p := c.Query("page"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 0 {
			c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
				Message: "Invalid data",
				Errors: []models.ErrorDescription{
					{Field: "page", Error: "must be >= 0"},
				},
			})
			return
		}
		page = v
	}

	if s := c.Query("size"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 1 || v > 100 {
			c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
				Message: "Invalid data",
				Errors: []models.ErrorDescription{
					{Field: "size", Error: "must be between 1 and 100"},
				},
			})
			return
		}
		size = v
	}

	resp, err := a.reservationService.ListHotels(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/v1/reservations
func (a *V1) HandleReservationList(c *gin.Context) {
	username := c.GetHeader("X-User-Name")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	reservations, err := a.reservationService.ListReservations(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, reservations)
}

// POST /api/v1/reservations
func (a *V1) HandleReservationCreate(c *gin.Context) {
	username := c.GetHeader("X-User-Name")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	var req models.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Message: "Invalid data",
			Errors: []models.ErrorDescription{
				{Field: "body", Error: err.Error()},
			},
		})
		return
	}

	if errs := validateCreateReservationRequest(req); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
			Message: "Invalid data",
			Errors:  errs,
		})
		return
	}

	resp, err := a.reservationService.CreateReservation(username, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrHotelNotFound):
			c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
				Message: "Invalid data",
				Errors: []models.ErrorDescription{
					{Field: "hotelUid", Error: "hotel not found"},
				},
			})
		case errors.Is(err, service.ErrInvalidData):
			c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
				Message: "Invalid data",
				Errors: []models.ErrorDescription{
					{Field: "body", Error: err.Error()},
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/v1/reservations/{reservationUid}
func (a *V1) HandleReservationGet(c *gin.Context) {
	username := c.GetHeader("X-User-Name")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	reservationUid := c.Param("reservationUid")
	if reservationUid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "reservationUid is required"})
		return
	}

	resp, err := a.reservationService.GetReservation(username, reservationUid)
	if err != nil {
		if errors.Is(err, service.ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Reservation not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DELETE /api/v1/reservations/{reservationUid}
func (a *V1) HandleReservationCancel(c *gin.Context) {
	username := c.GetHeader("X-User-Name")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return
	}

	reservationUid := c.Param("reservationUid")
	if reservationUid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "reservationUid is required"})
		return
	}

	err := a.reservationService.CancelReservation(username, reservationUid)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrReservationNotFound):
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Reservation not found"})
		case errors.Is(err, service.ErrReservationAlreadyCanceled):
			c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Reservation already canceled"})
		default:
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func validateCreateReservationRequest(req models.CreateReservationRequest) []models.ErrorDescription {
	errs := []models.ErrorDescription{}

	if req.HotelUid == "" {
		errs = append(errs, models.ErrorDescription{
			Field: "hotelUid",
			Error: "hotelUid is required",
		})
	}
	if req.StartDate == "" {
		errs = append(errs, models.ErrorDescription{
			Field: "startDate",
			Error: "startDate is required",
		})
	}
	if req.EndDate == "" {
		errs = append(errs, models.ErrorDescription{
			Field: "endDate",
			Error: "endDate is required",
		})
	}

	return errs
}
