package models

import "time"

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

type Hotel struct {
	HotelUid string `gorm:"column:hotel_uid;type:uuid;primaryKey" json:"hotelUid"`
	Name     string `gorm:"column:name;not null" json:"name"`
	Country  string `gorm:"column:country;not null" json:"country"`
	City     string `gorm:"column:city;not null" json:"city"`
	Address  string `gorm:"column:address;not null" json:"address"`
	Stars    int    `gorm:"column:stars;not null" json:"stars"`
	Price    int    `gorm:"column:price;not null" json:"price"`
}

type Reservation struct {
	ReservationUid string    `gorm:"column:reservation_uid;type:uuid;primaryKey" json:"reservationUid"`
	Username       string    `gorm:"column:username;not null;index" json:"username"`
	HotelUid       string    `gorm:"column:hotel_uid;type:uuid;not null;index" json:"hotelUid"`
	Hotel          Hotel     `gorm:"foreignKey:HotelUid;references:HotelUid" json:"hotel"`
	StartDate      time.Time `gorm:"column:start_date;type:date;not null" json:"startDate"`
	EndDate        time.Time `gorm:"column:end_date;type:date;not null" json:"endDate"`
	Status         string    `gorm:"column:status;not null" json:"status"`
	PaymentStatus  string    `gorm:"column:payment_status;not null" json:"paymentStatus"`
	Price          int       `gorm:"column:price;not null" json:"price"`
	Discount       int       `gorm:"column:discount;not null;default:0" json:"discount"`
}
