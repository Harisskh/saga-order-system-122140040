package orchestrator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OrderResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type PaymentResponse struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type ShippingResponse struct {
	ID        string    `json:"id"`
	OrderID   string    `json:"order_id"`
	Address   string    `json:"address"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateOrderRequest struct {
	UserID  string      `json:"user_id"`
	Items   []OrderItem `json:"items"`
	Address string      `json:"address"`
}

type Orchestrator struct {
	orderServiceURL    string
	paymentServiceURL  string
	shippingServiceURL string
	client             *http.Client
}

func NewOrchestrator(orderURL, paymentURL, shippingURL string) *Orchestrator {
	return &Orchestrator{
		orderServiceURL:    orderURL,
		paymentServiceURL:  paymentURL,
		shippingServiceURL: shippingURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (o *Orchestrator) CreateOrderSaga(req CreateOrderRequest) (string, error) {
	// Step 1: Create Order
	orderResp, err := o.createOrder(req)
	if err != nil {
		return "", fmt.Errorf("failed to create order: %w", err)
	}

	// Step 2: Process Payment
	paymentResp, err := o.processPayment(orderResp.ID, orderResp.TotalPrice)
	if err != nil {
		// Compensation: Cancel Order
		o.cancelOrder(orderResp.ID)
		return "", fmt.Errorf("failed to process payment: %w", err)
	}

	// Step 3: Start Shipping
	_, err = o.startShipping(orderResp.ID, req.Address)
	if err != nil {
		// Compensation: Refund Payment and Cancel Order
		o.refundPayment(orderResp.ID)
		o.cancelOrder(orderResp.ID)
		return "", fmt.Errorf("failed to start shipping: %w", err)
	}

	return orderResp.ID, nil
}

func (o *Orchestrator) createOrder(req CreateOrderRequest) (*OrderResponse, error) {
	orderReq := map[string]interface{}{
		"user_id": req.UserID,
		"items":   req.Items,
	}

	orderReqJSON, err := json.Marshal(orderReq)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Post(o.orderServiceURL+"/create-order", "application/json", bytes.NewBuffer(orderReqJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("order service returned status code %d", resp.StatusCode)
	}

	var orderResp OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, err
	}

	return &orderResp, nil
}

func (o *Orchestrator) processPayment(orderID string, amount float64) (*PaymentResponse, error) {
	paymentReq := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	}

	paymentReqJSON, err := json.Marshal(paymentReq)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Post(o.paymentServiceURL+"/process-payment", "application/json", bytes.NewBuffer(paymentReqJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("payment service returned status code %d", resp.StatusCode)
	}

	var paymentResp PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&paymentResp); err != nil {
		return nil, err
	}

	return &paymentResp, nil
}

func (o *Orchestrator) startShipping(orderID, address string) (*ShippingResponse, error) {
	shippingReq := map[string]interface{}{
		"order_id": orderID,
		"address":  address,
	}

	shippingReqJSON, err := json.Marshal(shippingReq)
	if err != nil {
		return nil, err
	}

	resp, err := o.client.Post(o.shippingServiceURL+"/start-shipping", "application/json", bytes.NewBuffer(shippingReqJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("shipping service returned status code %d", resp.StatusCode)
	}

	var shippingResp ShippingResponse
	if err := json.NewDecoder(resp.Body).Decode(&shippingResp); err != nil {
		return nil, err
	}

	return &shippingResp, nil
}

func (o *Orchestrator) cancelOrder(orderID string) error {
	cancelReq := map[string]interface{}{
		"order_id": orderID,
	}

	cancelReqJSON, err := json.Marshal(cancelReq)
	if err != nil {
		return err
	}

	resp, err := o.client.Post(o.orderServiceURL+"/cancel-order", "application/json", bytes.NewBuffer(cancelReqJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel order returned status code %d", resp.StatusCode)
	}

	return nil
}

func (o *Orchestrator) refundPayment(orderID string) error {
	refundReq := map[string]interface{}{
		"order_id": orderID,
	}

	refundReqJSON, err := json.Marshal(refundReq)
	if err != nil {
		return err
	}

	resp, err := o.client.Post(o.paymentServiceURL+"/refund-payment", "application/json", bytes.NewBuffer(refundReqJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refund payment returned status code %d", resp.StatusCode)
	}

	return nil
}

func (o *Orchestrator) cancelShipping(orderID string) error {
	cancelReq := map[string]interface{}{
		"order_id": orderID,
	}

	cancelReqJSON, err := json.Marshal(cancelReq)
	if err != nil {
		return err
	}

	resp, err := o.client.Post(o.shippingServiceURL+"/cancel-shipping", "application/json", bytes.NewBuffer(cancelReqJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel shipping returned status code %d", resp.StatusCode)
	}

	return nil
}