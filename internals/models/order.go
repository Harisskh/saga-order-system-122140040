package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Items     []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateOrderRequest struct {
	UserID string      `json:"user_id" binding:"required"`
	Items  []OrderItem `json:"items" binding:"required,min=1"`
}

type OrderResponse struct {
	ID         string      `json:"id"`
	UserID     string      `json:"user_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
}

func NewOrder(req CreateOrderRequest) Order {
	var totalPrice float64
	for _, item := range req.Items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	now := time.Now()
	return Order{
		ID:         uuid.New().String(),
		UserID:     req.UserID,
		Items:      req.Items,
		TotalPrice: totalPrice,
		Status:     OrderStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}