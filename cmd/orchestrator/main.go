package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"saga-go/models"
)

const (
	PORT            = 3000
	ORDER_SERVICE   = "http://localhost:3001"
	PAYMENT_SERVICE = "http://localhost:3002"
	SHIPPING_SERVICE = "http://localhost:3003"
)

func main() {
	http.HandleFunc("/create-order-saga", createOrderSagaHandler)

	fmt.Printf("Saga Orchestrator running on port %d\n", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil))
}

func createOrderSagaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var orderRequest models.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("Starting Create Order Saga for: %+v\n", orderRequest)

	// Step 1: Create Order
	fmt.Println("1. Creating order...")
	orderResp, err := createOrder(orderRequest)
	if err != nil {
		response := models.SagaResponse{
			Success:  false,
			Message:  "Saga failed at order creation step",
			ErrorMsg: err.Error(),
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}

	orderId := orderResp.OrderID
	fmt.Printf("Order created with ID: %s\n", orderId)

	// Step 2: Process Payment
	fmt.Println("2. Processing payment...")
	paymentReq := models.PaymentRequest{
		OrderID:       orderId,
		Amount:        orderRequest.Amount,
		PaymentMethod: orderRequest.PaymentMethod,
	}

	paymentResp, err := processPayment(paymentReq)
	if err != nil {
		// Payment failed, compensate
		fmt.Printf("Payment failed, starting compensation: %s\n", err.Error())
		
		// Attempt to cancel order
		cancelOrder(orderId)
		
		response := models.SagaResponse{
			Success:  false,
			Message:  "Saga failed at payment step. All operations have been compensated.",
			ErrorMsg: err.Error(),
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}

	fmt.Println("Payment processed successfully:", paymentResp)

	// Step 3: Start Shipping
	fmt.Println("3. Starting shipping...")
	shippingReq := models.ShippingRequest{
		OrderID: orderId,
		Address: orderRequest.ShippingAddress,
	}

	shippingResp, err := startShipping(shippingReq)
	if err != nil {
		// Shipping failed, compensate
		fmt.Printf("Shipping failed, starting compensation: %s\n", err.Error())
		
		// Cancel shipping (just in case it partially succeeded)
		cancelShipping(orderId)
		
		// Refund payment
		refundPayment(orderId)
		
		// Cancel order
		cancelOrder(orderId)
		
		response := models.SagaResponse{
			Success:  false,
			Message:  "Saga failed at shipping step. All operations have been compensated.",
			ErrorMsg: err.Error(),
		}
		sendResponse(w, response, http.StatusInternalServerError)
		return
	}

	fmt.Println("Shipping initiated successfully:", shippingResp)

	// Complete the order
	completeOrder(orderId)

	// All steps completed successfully
	response := models.SagaResponse{
		Success: true,
		Message: "Order saga completed successfully",
		Order: &models.OrderWithDetails{
			OrderID:  orderId,
			Status:   "COMPLETED",
			Payment:  paymentResp,
			Shipping: shippingResp,
		},
	}

	sendResponse(w, response, http.StatusOK)
}

func createOrder(orderReq models.OrderRequest) (*models.OrderResponse, error) {
	body, err := json.Marshal(orderReq)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/create-order", ORDER_SERVICE),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var orderResp models.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, err
	}

	if !orderResp.Success {
		return nil, fmt.Errorf("order creation failed: %s", orderResp.ErrorMsg)
	}

	return &orderResp, nil
}

func cancelOrder(orderId string) error {
	compReq := models.CompensationRequest{OrderID: orderId}
	body, err := json.Marshal(compReq)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/cancel-order", ORDER_SERVICE),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var orderResp models.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return err
	}

	if !orderResp.Success {
		return fmt.Errorf("order cancellation failed: %s", orderResp.ErrorMsg)
	}

	fmt.Printf("Order cancelled: %s\n", orderId)
	return nil
}

func completeOrder(orderId string) error {
	compReq := models.CompensationRequest{OrderID: orderId}
	body, err := json.Marshal(compReq)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/complete-order", ORDER_SERVICE),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func processPayment(paymentReq models.PaymentRequest) (*models.PaymentResponse, error) {
	body, err := json.Marshal(paymentReq)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/process-payment", PAYMENT_SERVICE),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var paymentResp models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, err
	}

	if !paymentResp.Success {
		return nil, fmt.Errorf("payment processing failed: %s", paymentResp.ErrorMsg)
	}

	return &paymentResp, nil
}

func refundPayment(orderId string) error {
	compReq := models.CompensationRequest{OrderID: orderId}
	body, err := json.Marshal(compReq)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/refund-payment", PAYMENT_SERVICE),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var paymentResp models.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return err
	}

	if !paymentResp.Success {
		return fmt.Errorf("payment refund failed: %s", paymentResp.ErrorMsg)
	}

	fmt.Printf("Payment refunded for order: %s\n", orderId)
	return nil
}

func startShipping(shippingReq models.ShippingRequest) (*models.ShippingResponse, error) {
	body, err := json.Marshal(shippingReq)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/start-shipping", SHIPPING_SERVICE),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var shippingResp models.ShippingResponse
	if err := json.NewDecoder(resp.Body).Decode(&shippingResp); err != nil {
		return nil, err
	}

	if !shippingResp.Success {
		return nil, fmt.Errorf("shipping initiation failed: %s", shippingResp.ErrorMsg)
	}

	return &shippingResp, nil
}

func cancelShipping(orderId string) error {
	compReq := models.CompensationRequest{OrderID: orderId}
	body, err := json.Marshal(compReq)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/cancel-shipping", SHIPPING_SERVICE),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var shippingResp models.ShippingResponse
	if err := json.NewDecoder(resp.Body).Decode(&shippingResp); err != nil {
		return err
	}

	if !shippingResp.Success {
		return fmt.Errorf("shipping cancellation failed: %s", shippingResp.ErrorMsg)
	}

	fmt.Printf("Shipping cancelled for order: %s\n", orderId)
	return nil
}

func sendResponse(w http.ResponseWriter, response interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}