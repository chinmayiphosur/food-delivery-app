package models

import "time"

// OrderStatus represents the current state of an order.
type OrderStatus string

const (
	StatusPlaced         OrderStatus = "PLACED"
	StatusConfirmed      OrderStatus = "CONFIRMED"
	StatusPreparing      OrderStatus = "PREPARING"
	StatusReadyForPickup OrderStatus = "READY_FOR_PICKUP"
	StatusPickedUp       OrderStatus = "PICKED_UP"
	StatusOutForDelivery OrderStatus = "OUT_FOR_DELIVERY"
	StatusDelivered      OrderStatus = "DELIVERED"
	StatusCancelled      OrderStatus = "CANCELLED"
)

// OrderItem represents a single item in an order.
type OrderItem struct {
	MenuItemID string  `json:"menu_item_id" bson:"menu_item_id"`
	Name       string  `json:"name" bson:"name"`
	Quantity   int     `json:"quantity" bson:"quantity"`
	Price      float64 `json:"price" bson:"price"`
}

// StatusChange records a single state transition in the order's history.
type StatusChange struct {
	FromStatus OrderStatus `json:"from_status" bson:"from_status"`
	ToStatus   OrderStatus `json:"to_status" bson:"to_status"`
	ChangedBy  string      `json:"changed_by" bson:"changed_by"`
	Role       Role        `json:"role" bson:"role"`
	Timestamp  time.Time   `json:"timestamp" bson:"timestamp"`
}

// Order represents a food delivery order.
type Order struct {
	ID              string         `json:"id" bson:"_id,omitempty"`
	CustomerID      string         `json:"customer_id" bson:"customer_id"`
	RestaurantID    string         `json:"restaurant_id" bson:"restaurant_id"`
	DriverID        string         `json:"driver_id,omitempty" bson:"driver_id,omitempty"`
	Items           []OrderItem    `json:"items" bson:"items"`
	TotalAmount     float64        `json:"total_amount" bson:"total_amount"`
	Status          OrderStatus    `json:"status" bson:"status"`
	StatusHistory   []StatusChange `json:"status_history" bson:"status_history"`
	DeliveryAddress string         `json:"delivery_address" bson:"delivery_address"`
	CreatedAt       time.Time      `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at" bson:"updated_at"`
}

// UpdateStatusRequest is the payload for updating order status.
type UpdateStatusRequest struct {
	Status   OrderStatus `json:"status"`
	DriverID string      `json:"driver_id,omitempty"`
}
