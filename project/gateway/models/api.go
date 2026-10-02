package models

import "fmt"

// ---------- статусы лояльности ----------

type LoyaltyStatus string

const (
	StatusBronze LoyaltyStatus = "BRONZE"
	StatusSilver LoyaltyStatus = "SILVER"
	StatusGold   LoyaltyStatus = "GOLD"
)

// ---------- общие ошибки ----------

type ErrorDescription struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Message string             `json:"message"`
	Errors  []ErrorDescription `json:"errors"`
}

// ---------- Loyalty Service ----------

type LoyaltyInfoResponse struct {
	Status           LoyaltyStatus `json:"status"`
	Discount         int           `json:"discount"`
	ReservationCount int           `json:"reservationCount"`
}

type LoyaltyRequest struct {
	Username         string        `json:"username"`
	ReservationCount int           `json:"reservationCount"`
	Status           LoyaltyStatus `json:"status"`
	Discount         int           `json:"discount"`
}

// ---------- Payment Service ----------

type PaymentRequest struct {
	Price int `json:"price"`
}

type PaymentResponse struct {
	PaymentUID string `json:"paymentUid"`
	Status     string `json:"status"`
	Price      int    `json:"price"`
}

type PaymentInfo struct {
	Status string `json:"status"`
	Price  int    `json:"price"`
}

// ---------- Reservation Service ----------

type PaginationResponse struct {
	Page          int             `json:"page"`
	PageSize      int             `json:"pageSize"`
	TotalElements int64           `json:"totalElements"`
	Items         []HotelResponse `json:"items"`
}

type HotelResponse struct {
	HotelUid string `json:"hotelUid"`
	Name     string `json:"name"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Stars    int    `json:"stars"`
	Price    int    `json:"price"`
}

type HotelInfo struct {
	HotelUid    string `json:"hotelUid"`
	Name        string `json:"name"`
	FullAddress string `json:"fullAddress"`
	Stars       int    `json:"stars"`
}

type ReservationResponse struct {
	ReservationUid string      `json:"reservationUid"`
	Hotel          HotelInfo   `json:"hotel"`
	StartDate      string      `json:"startDate"`
	EndDate        string      `json:"endDate"`
	Status         string      `json:"status"`
	Payment        PaymentInfo `json:"payment"`
}

// CreateReservationRequest уходит в reservation-service — там ждут уже
// посчитанные discount/price/status/paymentStatus.
type CreateReservationRequest struct {
	HotelUid      string `json:"hotelUid"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	Discount      int    `json:"discount"`
	Price         int    `json:"price"`
	Status        string `json:"status"`
	PaymentStatus string `json:"paymentStatus"`
}

type CreateReservationResponse struct {
	ReservationUid string      `json:"reservationUid"`
	HotelUid       string      `json:"hotelUid"`
	StartDate      string      `json:"startDate"`
	EndDate        string      `json:"endDate"`
	Discount       int         `json:"discount"`
	Status         string      `json:"status"`
	Payment        PaymentInfo `json:"payment"`
}

// MeResponse — ответ /api/v1/me
type MeResponse struct {
	Reservations []ReservationResponse `json:"reservations"`
	Loyalty      LoyaltyInfoResponse   `json:"loyalty"`
}

// ---------- константы статусов ----------

const (
	ReservationStatusPaid     = "PAID"
	ReservationStatusReserved = "RESERVED"
	ReservationStatusCanceled = "CANCELED"

	PaymentStatusPaid     = "PAID"
	PaymentStatusReversed = "REVERSED"
	PaymentStatusCanceled = "CANCELED"
)

// RemoteError — ошибка обращения к внешнему сервису.
type RemoteError struct {
	Service string
	Status  int
}

func (e *RemoteError) Error() string {
	return fmt.Sprintf("%s service returned status %d", e.Service, e.Status)
}
