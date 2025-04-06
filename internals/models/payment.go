package models

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusSuccess  PaymentStatus = "SUCCESS"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
)

type Payment struct {
	ID        string        `json:"id"`
	OrderID   string        `json:"order_id"`
	Amount    float64       `json:"amount"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type ProcessPaymentRequest struct {
	OrderID string  `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
}

type PaymentResponse struct {
	ID        string        `json:"id"`
	OrderID   string        `json:"order_id"`
	Amount    float64       `json:"amount"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
}

func NewPayment(req ProcessPaymentRequest) Payment {
	now := time.Now()
	return Payment{
		ID:        uuid.New().String(),
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Status:    PaymentStatusSuccess, // Default to success for simplicity
		CreatedAt: now,
		UpdatedAt: now,
	}
}