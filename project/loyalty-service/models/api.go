package models

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
