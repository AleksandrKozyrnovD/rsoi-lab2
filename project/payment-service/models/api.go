package models

type PaymentRequest struct {
	Price int `json:"price" binding:"required,gt=0"`
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
