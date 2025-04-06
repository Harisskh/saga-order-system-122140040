package payment

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"saga-order-system/internal/models"
)

type Service struct {
	payments map[string]models.Payment
	mu       sync.RWMutex
	// Untuk simulasi kegagalan pembayaran
	failNextPayment bool
}

func NewService() *Service {
	return &Service{
		payments:       make(map[string]models.Payment),
		failNextPayment: false,
	}
}

func (s *Service) SetupRoutes(router *gin.Engine) {
	router.POST("/process-payment", s.ProcessPayment)
	router.POST("/refund-payment", s.RefundPayment)
	router.GET("/payments/:id", s.GetPayment)
	router.POST("/set-fail-next-payment", s.SetFailNextPayment)
}

func (s *Service) SetFailNextPayment(c *gin.Context) {
	var req struct {
		Fail bool `json:"fail"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.failNextPayment = req.Fail
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Set fail next payment to %v", req.Fail)})
}

func (s *Service) ProcessPayment(c *gin.Context) {
	var req models.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simulasi kegagalan pembayaran
	if s.failNextPayment {
		s.failNextPayment = false
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment processing failed"})
		return
	}

	payment := models.NewPayment(req)

	s.mu.Lock()
	s.payments[payment.ID] = payment
	s.mu.Unlock()

	c.JSON(http.StatusCreated, models.PaymentResponse{
		ID:        payment.ID,
		OrderID:   payment.OrderID,
		Amount:    payment.Amount,
		Status:    payment.Status,
		CreatedAt: payment.CreatedAt,
	})
}

func (s *Service) RefundPayment(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var paymentFound bool
	for id, payment := range s.payments {
		if payment.OrderID == req.OrderID {
			payment.Status = models.PaymentStatusRefunded
			s.payments[id] = payment
			paymentFound = true
			break
		}
	}

	if !paymentFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found for the given order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Payment for order %s refunded successfully", req.OrderID)})
}

func (s *Service) GetPayment(c *gin.Context) {
	paymentID := c.Param("id")

	s.mu.RLock()
	payment, exists := s.payments[paymentID]
	s.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, models.PaymentResponse{
		ID:        payment.ID,
		OrderID:   payment.OrderID,
		Amount:    payment.Amount,
		Status:    payment.Status,
		CreatedAt: payment.CreatedAt,
	})
}