package models

// MenuItem represents a dish on a restaurant's menu.
type MenuItem struct {
	ID           string  `json:"id" bson:"_id,omitempty"`
	RestaurantID string  `json:"restaurant_id" bson:"restaurant_id"`
	Name         string  `json:"name" bson:"name"`
	Description  string  `json:"description" bson:"description"`
	Price        float64 `json:"price" bson:"price"`
	Category     string  `json:"category" bson:"category"`
	Available    bool    `json:"available" bson:"available"`
	ImageURL     string  `json:"image_url,omitempty" bson:"image_url,omitempty"`
}

// CreateMenuItemRequest is the payload for adding a menu item.
type CreateMenuItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	ImageURL    string  `json:"image_url,omitempty"`
}

// OrderItemRequest is used by customers to order from a menu.
type OrderItemRequest struct {
	MenuItemID string `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
}

// CreateOrderFromMenuRequest is the payload for placing an order from a restaurant's menu.
type CreateOrderFromMenuRequest struct {
	RestaurantID    string             `json:"restaurant_id"`
	Items           []OrderItemRequest `json:"items"`
	DeliveryAddress string             `json:"delivery_address"`
	PaymentMethod   string             `json:"payment_method"`
}
