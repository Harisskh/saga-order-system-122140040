package order

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"saga-order-system/internal/models"
)

type Service struct {
	orders map[string]models.Order
	mu     sync.RWMutex
}

func NewService() *Service {
	return &Service{
		orders: make(map[string]models.Order),
	}
}

func (s *Service) SetupRoutes(router *gin.Engine) {
	router.POST("/create-order", s.CreateOrder)
	router.POST("/cancel-order", s.CancelOrder)
	router.GET("/orders/:id", s.GetOrder)
}

func (s *Service) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := models.NewOrder(req)

	s.mu.Lock()
	s.orders[order.ID] = order
	s.mu.Unlock()

	c.JSON(http.StatusCreated, models.OrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		Items:      order.Items,
		TotalPrice: order.TotalPrice,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
	})
}

func (s *Service) CancelOrder(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[req.OrderID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.Status = models.OrderStatusCancelled
	s.orders[req.OrderID] = order

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Order %s cancelled successfully", req.OrderID)})
}

func (s *Service) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	s.mu.RLock()
	order, exists := s.orders[orderID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, models.OrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		Items:      order.Items,
		TotalPrice: order.TotalPrice,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
	})
}

func (s *Service) CompleteOrder(orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return fmt.Errorf("order not found")
	}

	order.Status = models.OrderStatusCompleted
	s.orders[orderID] = order

	return nil
}