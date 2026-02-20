package handlers

import (
	"encoding/json"
	"food-delivery-api/db"
	"food-delivery-api/models"
	"food-delivery-api/statemachine"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// OrderHandler handles order-related HTTP requests.
type OrderHandler struct {
	Store *db.Store
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(store *db.Store) *OrderHandler {
	return &OrderHandler{Store: store}
}

// CreateOrder handles POST /api/orders
// Customers select dishes from a restaurant's menu. Items are looked up by menu_item_id.
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(ContextKeyUserRole).(string)
	userID := r.Context().Value(ContextKeyUserID).(string)

	if models.Role(role) != models.RoleCustomer {
		respondError(w, http.StatusForbidden, "Only customers can create orders")
		return
	}

	var req models.CreateOrderFromMenuRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RestaurantID == "" {
		respondError(w, http.StatusBadRequest, "restaurant_id is required")
		return
	}
	if len(req.Items) == 0 {
		respondError(w, http.StatusBadRequest, "At least one item is required")
		return
	}
	if req.DeliveryAddress == "" {
		respondError(w, http.StatusBadRequest, "delivery_address is required")
		return
	}

	// Verify the restaurant exists.
	restaurant, err := h.Store.GetUser(req.RestaurantID)
	if err != nil || restaurant.Role != models.RoleRestaurant {
		respondError(w, http.StatusBadRequest, "Invalid restaurant_id")
		return
	}

	// Look up each menu item and build order items.
	var orderItems []models.OrderItem
	var total float64
	for _, ri := range req.Items {
		if ri.Quantity <= 0 {
			respondError(w, http.StatusBadRequest, "Quantity must be at least 1")
			return
		}
		menuItem, err := h.Store.GetMenuItem(ri.MenuItemID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Menu item not found: "+ri.MenuItemID)
			return
		}
		if menuItem.RestaurantID != req.RestaurantID {
			respondError(w, http.StatusBadRequest, "Menu item "+menuItem.Name+" does not belong to this restaurant")
			return
		}
		if !menuItem.Available {
			respondError(w, http.StatusBadRequest, "Menu item '"+menuItem.Name+"' is currently unavailable")
			return
		}
		orderItems = append(orderItems, models.OrderItem{
			MenuItemID: menuItem.ID,
			Name:       menuItem.Name,
			Quantity:   ri.Quantity,
			Price:      menuItem.Price,
		})
		total += menuItem.Price * float64(ri.Quantity)
	}

	now := time.Now()
	order := &models.Order{
		ID:              uuid.New().String(),
		CustomerID:      userID,
		RestaurantID:    req.RestaurantID,
		Items:           orderItems,
		TotalAmount:     total,
		Status:          models.StatusPlaced,
		DeliveryAddress: req.DeliveryAddress,
		StatusHistory: []models.StatusChange{
			{
				FromStatus: "",
				ToStatus:   models.StatusPlaced,
				ChangedBy:  userID,
				Role:       models.RoleCustomer,
				Timestamp:  now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.Store.SaveOrder(order); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save order")
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

// GetOrder handles GET /api/orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	order, err := h.Store.GetOrder(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, order)
}

// ListOrders handles GET /api/orders
// Supports optional ?status= query parameter for filtering.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	statusFilter := models.OrderStatus(r.URL.Query().Get("status"))
	orders, err := h.Store.ListOrders(statusFilter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}
	respondJSON(w, http.StatusOK, orders)
}

// UpdateOrderStatus handles PATCH /api/orders/{id}/status
// Validates the transition using the state machine and role permissions.
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	role := r.Context().Value(ContextKeyUserRole).(string)
	userID := r.Context().Value(ContextKeyUserID).(string)

	order, err := h.Store.GetOrder(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate the state transition using the state machine.
	if err := statemachine.ValidateTransition(order.Status, req.Status, models.Role(role)); err != nil {
		// Determine if it's a role permission issue (403) or invalid transition (400).
		allRoleErr := statemachine.ValidateTransition(order.Status, req.Status, models.RoleCustomer)
		if allRoleErr != nil {
			allRoleErr = statemachine.ValidateTransition(order.Status, req.Status, models.RoleRestaurant)
		}
		if allRoleErr != nil {
			allRoleErr = statemachine.ValidateTransition(order.Status, req.Status, models.RoleDriver)
		}

		if allRoleErr == nil {
			respondError(w, http.StatusForbidden, err.Error())
		} else {
			respondError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// Assign driver if transitioning to PICKED_UP.
	if req.Status == models.StatusPickedUp && order.DriverID == "" {
		order.DriverID = userID
	}

	// Record the status change.
	now := time.Now()
	order.StatusHistory = append(order.StatusHistory, models.StatusChange{
		FromStatus: order.Status,
		ToStatus:   req.Status,
		ChangedBy:  userID,
		Role:       models.Role(role),
		Timestamp:  now,
	})

	order.Status = req.Status
	order.UpdatedAt = now
	if err := h.Store.SaveOrder(order); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update order")
		return
	}

	respondJSON(w, http.StatusOK, order)
}

// GetOrderHistory handles GET /api/orders/{id}/history
func (h *OrderHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	order, err := h.Store.GetOrder(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, order.StatusHistory)
}

// GetAllowedTransitions handles GET /api/orders/{id}/transitions
func (h *OrderHandler) GetAllowedTransitions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	role := r.Context().Value(ContextKeyUserRole).(string)

	order, err := h.Store.GetOrder(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	transitions := statemachine.GetAllowedTransitions(order.Status, models.Role(role))
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"current_status":      order.Status,
		"allowed_transitions": transitions,
	})
}
