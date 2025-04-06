package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"saga-order-system/internal/services/order"
)

func main() {
	r := gin.Default()

	service := order.NewService()
	service.SetupRoutes(r)

	log.Println("Order service starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start order service: %v", err)
	}
}