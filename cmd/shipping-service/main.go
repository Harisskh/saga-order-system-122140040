package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"saga-order-system/internal/services/shipping"
)

func main() {
	r := gin.Default()

	service := shipping.NewService()
	service.SetupRoutes(r)

	log.Println("Shipping service starting on :8083")
	if err := r.Run(":8083"); err != nil {
		log.Fatalf("Failed to start shipping service: %v", err)
	}
}