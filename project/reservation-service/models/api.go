package models

// PaginationResponse — ответ со списком отелей.
type PaginationResponse struct {
	Page          int             `json:"page"`
	PageSize      int             `json:"pageSize"`
	TotalElements int64           `json:"totalElements"`
	Items         []HotelResponse `json:"items"`
}

// HotelResponse — отель в списке.
type HotelResponse struct {
	HotelUid string `json:"hotelUid"`
	Name     string `json:"name"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Address  string `json:"address"`
	Stars    int    `json:"stars"`
	Price    int    `json:"price"`
}

// HotelInfo — краткая информация об отеле внутри брони.
type HotelInfo struct {
	HotelUid    string `json:"hotelUid"`
	Name        string `json:"name"`
	FullAddress string `json:"fullAddress"`
	Stars       int    `json:"stars"`
}

// ReservationResponse — бронь.
type ReservationResponse struct {
	ReservationUid string      `json:"reservationUid"`
	Hotel          HotelInfo   `json:"hotel"`
	StartDate      string      `json:"startDate"`
	EndDate        string      `json:"endDate"`
	Status         string      `json:"status"`
	Payment        PaymentInfo `json:"payment"`
}

// CreateReservationRequest — запрос на создание брони.
// Приходит из API Gateway уже с рассчитанными скидкой, ценой и статусами.
type CreateReservationRequest struct {
	HotelUid      string `json:"hotelUid"`
	StartDate     string `json:"startDate"`
	EndDate       string `json:"endDate"`
	Discount      int    `json:"discount"`
	Price         int    `json:"price"`
	Status        string `json:"status"`
	PaymentStatus string `json:"paymentStatus"`
}

// CreateReservationResponse — ответ на создание брони.
type CreateReservationResponse struct {
	ReservationUid string      `json:"reservationUid"`
	HotelUid       string      `json:"hotelUid"`
	StartDate      string      `json:"startDate"`
	EndDate        string      `json:"endDate"`
	Discount       int         `json:"discount"`
	Status         string      `json:"status"`
	Payment        PaymentInfo `json:"payment"`
}

// PaymentInfo — информация об оплате.
type PaymentInfo struct {
	Status string `json:"status"`
	Price  int    `json:"price"`
}

const (
	ReservationStatusPaid     = "PAID"
	ReservationStatusReserved = "RESERVED"
	ReservationStatusCanceled = "CANCELED"

	PaymentStatusPaid     = "PAID"
	PaymentStatusReversed = "REVERSED"
	PaymentStatusCanceled = "CANCELED"
)
