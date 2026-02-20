package handlers

import (
	"encoding/json"
	"food-delivery-api/db"
	"food-delivery-api/models"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	Store *db.Store
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(store *db.Store) *UserHandler {
	return &UserHandler{Store: store}
}

// RegisterUser handles POST /api/users
// Creates a new user with the specified name and role.
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if !req.Role.IsValid() {
		respondError(w, http.StatusBadRequest, "Role must be one of: customer, restaurant, driver")
		return
	}

	user := &models.User{
		ID:   uuid.New().String(),
		Name: req.Name,
		Role: req.Role,
	}
	if err := h.Store.SaveUser(user); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save user")
		return
	}

	respondJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /api/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	user, err := h.Store.GetUser(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// ListUsers handles GET /api/users
// Supports optional ?role= query parameter for filtering.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	roleFilter := models.Role(r.URL.Query().Get("role"))
	users, err := h.Store.ListUsers(roleFilter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	respondJSON(w, http.StatusOK, users)
}
