package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"gateway/models"
)

type V1 struct {
	loyaltyServiceURL     string
	reservationServiceURL string
	paymentServiceURL     string
	httpClient            *http.Client
}

func NewV1(loyaltyServiceURL, reservationServiceURL, paymentServiceURL string) *V1 {
	return &V1{
		loyaltyServiceURL:     loyaltyServiceURL,
		reservationServiceURL: reservationServiceURL,
		paymentServiceURL:     paymentServiceURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ============================================================
// helpers
// ============================================================

// getUserName достаёт X-User-Name. Если заголовка нет — пишет 400 и false.
func getUserName(c *gin.Context) (string, bool) {
	username := c.GetHeader("X-User-Name")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "X-User-Name header is required"})
		return "", false
	}
	return username, true
}

// doJSON делает HTTP-запрос. Если out != nil — декодирует тело.
func (a *V1) doJSON(method, rawURL string, body any, headers map[string]string, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if out != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, out); err != nil {
				return resp.StatusCode, err
			}
		}
	}
	return resp.StatusCode, nil
}

// getLoyalty тянет лояльность; если сервис отдаёт 404 — создаёт запись.
func (a *V1) getLoyalty(username string) (*models.LoyaltyInfoResponse, error) {
	var loyalty models.LoyaltyInfoResponse

	status, err := a.doJSON(
		http.MethodGet,
		a.loyaltyServiceURL+"/api/v1/loyalty/"+url.PathEscape(username),
		nil, nil, &loyalty,
	)
	if err != nil {
		return nil, err
	}

	if status == http.StatusOK {
		return &loyalty, nil
	}

	if status == http.StatusNotFound {
		createReq := models.LoyaltyRequest{
			Username:         username,
			ReservationCount: 0,
			Status:           models.StatusBronze,
			Discount:         5,
		}
		var created models.LoyaltyInfoResponse
		status, err = a.doJSON(http.MethodPost, a.loyaltyServiceURL+"/api/v1/loyalty", createReq, nil, &created)
		if err != nil {
			return nil, err
		}
		if status >= 400 {
			return nil, &models.RemoteError{Service: "loyalty", Status: status}
		}
		return &created, nil
	}

	return nil, &models.RemoteError{Service: "loyalty", Status: status}
}

// findHotel ищет отель по uid среди всех отелей reservation-service.
func (a *V1) findHotel(hotelUid string) (*models.HotelResponse, error) {
	page := 1
	size := 100

	for {
		var resp models.PaginationResponse
		rawURL := fmt.Sprintf("%s/api/v1/hotels?page=%d&size=%d",
			a.reservationServiceURL, page, size)

		status, err := a.doJSON(http.MethodGet, rawURL, nil, nil, &resp)
		if err != nil {
			return nil, err
		}
		if status >= 400 {
			return nil, &models.RemoteError{Service: "reservation", Status: status}
		}

		for i := range resp.Items {
			if resp.Items[i].HotelUid == hotelUid {
				return &resp.Items[i], nil
			}
		}

		// дошли до конца пагинации
		if len(resp.Items) < size {
			break
		}
		page++
	}
	return nil, nil
}

// nightsBetween считает количество ночей между двумя датами YYYY-MM-DD.
func nightsBetween(start, end string) (int, error) {
	s, err := time.Parse("2006-01-02", start)
	if err != nil {
		return 0, err
	}
	e, err := time.Parse("2006-01-02", end)
	if err != nil {
		return 0, err
	}
	return int(e.Sub(s).Hours() / 24), nil
}

// writeValidationError — сокращение для частого ответа.
func writeValidationError(c *gin.Context, errs ...models.ErrorDescription) {
	c.JSON(http.StatusBadRequest, models.ValidationErrorResponse{
		Message: "Invalid data",
		Errors:  errs,
	})
}

// ============================================================
// GET /api/v1/hotels?page=&size=
// ============================================================

func (a *V1) HandleGetHotels(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")

	p, err := strconv.Atoi(page)
	if err != nil || p < 0 {
		writeValidationError(c, models.ErrorDescription{Field: "page", Error: "must be a non-negative integer"})
		return
	}
	s, err := strconv.Atoi(size)
	if err != nil || s < 1 || s > 100 {
		writeValidationError(c, models.ErrorDescription{Field: "size", Error: "must be between 1 and 100"})
		return
	}

	var resp models.PaginationResponse
	status, err := a.doJSON(
		http.MethodGet,
		fmt.Sprintf("%s/api/v1/hotels?page=%d&size=%d", a.reservationServiceURL, p, s),
		nil, nil, &resp,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if status >= 400 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("reservation service returned status %d", status),
		})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ============================================================
// GET /api/v1/me
// ============================================================

func (a *V1) HandleGetMe(c *gin.Context) {
	username, ok := getUserName(c)
	if !ok {
		return
	}

	// 1) брони пользователя
	var reservations []models.ReservationResponse
	status, err := a.doJSON(
		http.MethodGet,
		a.reservationServiceURL+"/api/v1/reservations",
		nil, map[string]string{"X-User-Name": username}, &reservations,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if status >= 400 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("reservation service returned status %d", status),
		})
		return
	}
	if reservations == nil {
		reservations = []models.ReservationResponse{}
	}

	// 2) лояльность (создаём если ещё нет)
	loyalty, err := a.getLoyalty(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MeResponse{
		Reservations: reservations,
		Loyalty:      *loyalty,
	})
}

// ============================================================
// GET /api/v1/reservations
// ============================================================

func (a *V1) HandleGetReservations(c *gin.Context) {
	username, ok := getUserName(c)
	if !ok {
		return
	}

	var reservations []models.ReservationResponse
	status, err := a.doJSON(
		http.MethodGet,
		a.reservationServiceURL+"/api/v1/reservations",
		nil, map[string]string{"X-User-Name": username}, &reservations,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if status >= 400 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("reservation service returned status %d", status),
		})
		return
	}
	if reservations == nil {
		reservations = []models.ReservationResponse{}
	}
	c.JSON(http.StatusOK, reservations)
}

// ============================================================
// POST /api/v1/reservations
// ============================================================

func (a *V1) HandleCreateReservation(c *gin.Context) {
	username, ok := getUserName(c)
	if !ok {
		return
	}

	var req struct {
		HotelUid  string `json:"hotelUid"`
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, models.ErrorDescription{Field: "body", Error: err.Error()})
		return
	}

	// валидация входных полей
	errs := []models.ErrorDescription{}
	if req.HotelUid == "" {
		errs = append(errs, models.ErrorDescription{Field: "hotelUid", Error: "hotelUid is required"})
	}
	if req.StartDate == "" {
		errs = append(errs, models.ErrorDescription{Field: "startDate", Error: "startDate is required"})
	}
	if req.EndDate == "" {
		errs = append(errs, models.ErrorDescription{Field: "endDate", Error: "endDate is required"})
	}
	if len(errs) > 0 {
		writeValidationError(c, errs...)
		return
	}

	// 1) лояльность
	loyalty, err := a.getLoyalty(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	// 2) отель
	hotel, err := a.findHotel(req.HotelUid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if hotel == nil {
		writeValidationError(c, models.ErrorDescription{Field: "hotelUid", Error: "hotel not found"})
		return
	}

	// 3) цена
	nights, err := nightsBetween(req.StartDate, req.EndDate)
	if err != nil {
		writeValidationError(c, models.ErrorDescription{Field: "startDate", Error: "invalid date format, expected YYYY-MM-DD"})
		return
	}
	if nights <= 0 {
		writeValidationError(c, models.ErrorDescription{Field: "endDate", Error: "endDate must be after startDate"})
		return
	}
	totalPrice := hotel.Price * nights
	discountedPrice := totalPrice * (100 - loyalty.Discount) / 100

	// 4) платёж
	var payment models.PaymentResponse
	status, err := a.doJSON(
		http.MethodPost,
		a.paymentServiceURL+"/api/v1/payment",
		models.PaymentRequest{Price: discountedPrice}, nil, &payment,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if status >= 400 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("payment service returned status %d", status),
		})
		return
	}

	// 5) бронь
	createReq := models.CreateReservationRequest{
		HotelUid:      req.HotelUid,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Discount:      loyalty.Discount,
		Price:         discountedPrice,
		Status:        models.ReservationStatusPaid,
		PaymentStatus: payment.Status,
	}
	var created models.CreateReservationResponse
	status, err = a.doJSON(
		http.MethodPost,
		a.reservationServiceURL+"/api/v1/reservations",
		createReq, map[string]string{"X-User-Name": username}, &created,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	if status >= 400 {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("reservation service returned status %d", status),
		})
		return
	}

	// 6) инкремент лояльности (best effort)
	var updated models.LoyaltyInfoResponse
	_, _ = a.doJSON(
		http.MethodPost,
		a.loyaltyServiceURL+"/api/v1/loyalty/"+url.PathEscape(username)+"/reservations",
		nil, nil, &updated,
	)

	c.JSON(http.StatusOK, created)
}

// ============================================================
// GET /api/v1/reservations/{reservationUid}
// ============================================================

func (a *V1) HandleGetReservationByUid(c *gin.Context) {
	username, ok := getUserName(c)
	if !ok {
		return
	}
	reservationUid := c.Param("reservationUid")
	if reservationUid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "reservationUid is required"})
		return
	}

	var reservation models.ReservationResponse
	status, err := a.doJSON(
		http.MethodGet,
		a.reservationServiceURL+"/api/v1/reservations/"+url.PathEscape(reservationUid),
		nil, map[string]string{"X-User-Name": username}, &reservation,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	switch {
	case status == http.StatusNotFound:
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Reservation not found"})
		return
	case status >= 400:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("reservation service returned status %d", status),
		})
		return
	}

	c.JSON(http.StatusOK, reservation)
}

// ============================================================
// DELETE /api/v1/reservations/{reservationUid}
// ============================================================

func (a *V1) HandleCancelReservation(c *gin.Context) {
	username, ok := getUserName(c)
	if !ok {
		return
	}
	reservationUid := c.Param("reservationUid")
	if reservationUid == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "reservationUid is required"})
		return
	}

	// 1) отменить бронь
	status, err := a.doJSON(
		http.MethodDelete,
		a.reservationServiceURL+"/api/v1/reservations/"+url.PathEscape(reservationUid),
		nil, map[string]string{"X-User-Name": username}, nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}
	switch {
	case status == http.StatusNotFound:
		c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Reservation not found"})
		return
	case status == http.StatusBadRequest:
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Reservation already canceled"})
		return
	case status >= 400:
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: fmt.Sprintf("reservation service returned status %d", status),
		})
		return
	}

	// 2) декремент лояльности (best effort)
	var updated models.LoyaltyInfoResponse
	_, _ = a.doJSON(
		http.MethodDelete,
		a.loyaltyServiceURL+"/api/v1/loyalty/"+url.PathEscape(username)+"/reservations",
		nil, nil, &updated,
	)

	c.Status(http.StatusNoContent)
}

// ============================================================
// GET /api/v1/loyalty
// ============================================================

func (a *V1) HandleGetLoyalty(c *gin.Context) {
	username, ok := getUserName(c)
	if !ok {
		return
	}

	loyalty, err := a.getLoyalty(username)
	if err != nil {
		// 404 от loyalty превращаем в 500 только если это не наш сценарий
		var remote *models.RemoteError
		if errors.As(err, &remote) && remote.Status == http.StatusNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Loyalty info not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, loyalty)
}
