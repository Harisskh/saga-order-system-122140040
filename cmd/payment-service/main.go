package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"saga-order-system/internal/services/payment"
)

func main() {
	r := gin.Default()

	service := payment.NewService()
	service.SetupRoutes(r)

	log.Println("Payment service starting on :8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Failed to start payment service: %v", err)
	}
}