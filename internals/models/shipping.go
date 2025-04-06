package models

import (
	"time"

	"github.com/google/uuid"
)

type ShippingStatus string

const (
	ShippingStatusPending   ShippingStatus = "PENDING"
	ShippingStatusShipped   ShippingStatus = "SHIPPED"
	ShippingStatusCancelled ShippingStatus = "CANCELLED"
)

type Shipping struct {
	ID        string         `json:"id"`
	OrderID   string         `json:"order_id"`
	Address   string         `json:"address"`
	Status    ShippingStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type StartShippingRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type ShippingResponse struct {
	ID        string         `json:"id"`
	OrderID   string         `json:"order_id"`
	Address   string         `json:"address"`
	Status    ShippingStatus `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
}

func NewShipping(req StartShippingRequest) Shipping {
	now := time.Now()
	return Shipping{
		ID:        uuid.New().String(),
		OrderID:   req.OrderID,
		Address:   req.Address,
		Status:    ShippingStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
}