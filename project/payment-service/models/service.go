package models

import "github.com/google/uuid"

const (
	PaymentStatusPaid     = "PAID"
	PaymentStatusCanceled = "CANCELED"
)

// GORM-модель таблицы payment
type Payment struct {
	ID         int       `gorm:"primaryKey;column:id"`
	PaymentUID uuid.UUID `gorm:"column:payment_uid;type:uuid;uniqueIndex;not null"`
	Status     string    `gorm:"column:status;type:varchar(20);not null;check:status IN ('PAID','CANCELED')"`
	Price      int       `gorm:"column:price;not null"`
}

// Ошибки
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
