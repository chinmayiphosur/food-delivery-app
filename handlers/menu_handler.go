package handlers

import (
	"encoding/json"
	"food-delivery-api/db"
	"food-delivery-api/models"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// MenuHandler handles menu-related HTTP requests.
type MenuHandler struct {
	Store *db.Store
}

// NewMenuHandler creates a new MenuHandler.
func NewMenuHandler(store *db.Store) *MenuHandler {
	return &MenuHandler{Store: store}
}

// AddMenuItem handles POST /api/restaurants/{id}/menu
// Only the restaurant owner can add items to their menu.
func (h *MenuHandler) AddMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID := vars["id"]

	role := r.Context().Value(ContextKeyUserRole).(string)
	userID := r.Context().Value(ContextKeyUserID).(string)

	if models.Role(role) != models.RoleRestaurant {
		respondError(w, http.StatusForbidden, "Only restaurants can manage menus")
		return
	}
	if userID != restaurantID {
		respondError(w, http.StatusForbidden, "You can only manage your own menu")
		return
	}

	var req models.CreateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Dish name is required")
		return
	}
	if req.Price <= 0 {
		respondError(w, http.StatusBadRequest, "Price must be greater than 0")
		return
	}
	if req.Category == "" {
		req.Category = "General"
	}

	item := &models.MenuItem{
		ID:           uuid.New().String(),
		RestaurantID: restaurantID,
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		Category:     req.Category,
		Available:    true,
		ImageURL:     req.ImageURL,
	}

	if err := h.Store.SaveMenuItem(item); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save menu item")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

// GetMenu handles GET /api/restaurants/{id}/menu
// Public endpoint — anyone can view a restaurant's menu.
func (h *MenuHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID := vars["id"]

	items, err := h.Store.ListMenuItems(restaurantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch menu")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// DeleteMenuItem handles DELETE /api/restaurants/{id}/menu/{itemId}
func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID := vars["id"]
	itemID := vars["itemId"]

	role := r.Context().Value(ContextKeyUserRole).(string)
	userID := r.Context().Value(ContextKeyUserID).(string)

	if models.Role(role) != models.RoleRestaurant || userID != restaurantID {
		respondError(w, http.StatusForbidden, "You can only manage your own menu")
		return
	}

	// Verify the item belongs to this restaurant.
	item, err := h.Store.GetMenuItem(itemID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Menu item not found")
		return
	}
	if item.RestaurantID != restaurantID {
		respondError(w, http.StatusForbidden, "Item does not belong to your restaurant")
		return
	}

	if err := h.Store.DeleteMenuItem(itemID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete menu item")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Menu item deleted"})
}
