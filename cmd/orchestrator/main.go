package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"saga-order-system/internal/orchestrator"
)

func main() {
	r := gin.Default()

	orch := orchestrator.NewOrchestrator(
		"http://localhost:8081", // Order Service
		"http://localhost:8082", // Payment Service
		"http://localhost:8083", // Shipping Service
	)

	r.POST("/create-order-saga", func(c *gin.Context) {
		var req orchestrator.CreateOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		orderID, err := orch.CreateOrderSaga(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "Order saga completed successfully",
			"order_id": orderID,
		})
	})

	log.Println("Orchestrator service starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}
}