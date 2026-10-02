package models

type LoyaltyStatus string

const (
	StatusBronze LoyaltyStatus = "BRONZE"
	StatusSilver LoyaltyStatus = "SILVER"
	StatusGold   LoyaltyStatus = "GOLD"
)

type Loyalty struct {
	ID               uint          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username         string        `gorm:"column:username;type:varchar(80);not null;uniqueIndex" json:"username"`
	ReservationCount int           `gorm:"column:reservation_count;not null;default:0" json:"reservation_count"`
	Status           LoyaltyStatus `gorm:"column:status;type:varchar(80);not null;default:'BRONZE';check:status IN ('BRONZE','SILVER','GOLD')" json:"status"`
	Discount         int           `gorm:"column:discount;not null" json:"discount"`
}

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
