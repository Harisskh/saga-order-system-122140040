package shipping

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"saga-order-system/internal/models"
)

type Service struct {
	shippings map[string]models.Shipping
	mu        sync.RWMutex
	// Untuk simulasi kegagalan pengiriman
	failNextShipping bool
}

func NewService() *Service {
	return &Service{
		shippings:        make(map[string]models.Shipping),
		failNextShipping: false,
	}
}

func (s *Service) SetupRoutes(router *gin.Engine) {
	router.POST("/start-shipping", s.StartShipping)
	router.POST("/cancel-shipping", s.CancelShipping)
	router.GET("/shippings/:id", s.GetShipping)
	router.POST("/set-fail-next-shipping", s.SetFailNextShipping)
}

func (s *Service) SetFailNextShipping(c *gin.Context) {
	var req struct {
		Fail bool `json:"fail"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.failNextShipping = req.Fail
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Set fail next shipping to %v", req.Fail)})
}

func (s *Service) StartShipping(c *gin.Context) {
	var req models.StartShippingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simulasi kegagalan pengiriman
	if s.failNextShipping {
		s.failNextShipping = false
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Shipping failed"})
		return
	}

	shipping := models.NewShipping(req)
	shipping.Status = models.ShippingStatusShipped

	s.mu.Lock()
	s.shippings[shipping.ID] = shipping
	s.mu.Unlock()

	c.JSON(http.StatusCreated, models.ShippingResponse{
		ID:        shipping.ID,
		OrderID:   shipping.OrderID,
		Address:   shipping.Address,
		Status:    shipping.Status,
		CreatedAt: shipping.CreatedAt,
	})
}

func (s *Service) CancelShipping(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var shippingFound bool
	for id, shipping := range s.shippings {
		if shipping.OrderID == req.OrderID {
			shipping.Status = models.ShippingStatusCancelled
			s.shippings[id] = shipping
			shippingFound = true
			break
		}
	}

	if !shippingFound {
		// Jika tidak ditemukan, kita anggap berhasil karena mungkin belum dibuat
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("No shipping found for order %s", req.OrderID)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Shipping for order %s cancelled successfully", req.OrderID)})
}

func (s *Service) GetShipping(c *gin.Context) {
	shippingID := c.Param("id")

	s.mu.RLock()
	shipping, exists := s.shippings[shippingID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shipping not found"})
		return
	}

	c.JSON(http.StatusOK, models.ShippingResponse{
		ID:        shipping.ID,
		OrderID:   shipping.OrderID,
		Address:   shipping.Address,
		Status:    shipping.Status,
		CreatedAt: shipping.CreatedAt,
	})
}