package statemachine

import (
	"fmt"
	"food-delivery-api/models"
)

// transition defines an allowed state change along with which roles may perform it.
type transition struct {
	To           models.OrderStatus
	AllowedRoles []models.Role
}

// transitionMap defines every valid transition from each state.
// This is the single source of truth for the order lifecycle.
var transitionMap = map[models.OrderStatus][]transition{
	models.StatusPlaced: {
		{To: models.StatusConfirmed, AllowedRoles: []models.Role{models.RoleRestaurant}},
		{To: models.StatusCancelled, AllowedRoles: []models.Role{models.RoleCustomer}},
	},
	models.StatusConfirmed: {
		{To: models.StatusPreparing, AllowedRoles: []models.Role{models.RoleRestaurant}},
		{To: models.StatusCancelled, AllowedRoles: []models.Role{models.RoleCustomer, models.RoleRestaurant}},
	},
	models.StatusPreparing: {
		{To: models.StatusReadyForPickup, AllowedRoles: []models.Role{models.RoleRestaurant}},
	},
	models.StatusReadyForPickup: {
		{To: models.StatusPickedUp, AllowedRoles: []models.Role{models.RoleDriver}},
	},
	models.StatusPickedUp: {
		{To: models.StatusOutForDelivery, AllowedRoles: []models.Role{models.RoleDriver}},
	},
	models.StatusOutForDelivery: {
		{To: models.StatusDelivered, AllowedRoles: []models.Role{models.RoleDriver}},
	},
	// Terminal states – no transitions allowed from DELIVERED or CANCELLED.
}

// ValidateTransition checks whether moving from the order's current status to
// newStatus is allowed, and whether the given role has permission to make 
// that transition.
//
// It returns nil on success, or a descriptive error explaining why the
// transition was denied:
//   - Invalid status value
//   - No transitions available from the current state (terminal state)
//   - The requested transition is not in the allowed list
//   - The caller's role does not have permission
func ValidateTransition(currentStatus models.OrderStatus, newStatus models.OrderStatus, role models.Role) error {
	// Check if the current state has any transitions at all.
	allowedTransitions, exists := transitionMap[currentStatus]
	if !exists {
		return fmt.Errorf("no transitions allowed from status '%s' (terminal state)", currentStatus)
	}

	// Look for the requested target status in the allowed transitions.
	for _, t := range allowedTransitions {
		if t.To == newStatus {
			// Found the transition – now check role permission.
			for _, allowedRole := range t.AllowedRoles {
				if allowedRole == role {
					return nil // Transition is valid and role is authorized.
				}
			}
			return fmt.Errorf(
				"role '%s' is not authorized to transition order from '%s' to '%s'",
				role, currentStatus, newStatus,
			)
		}
	}

	// Build a list of valid targets for a helpful error message.
	validTargets := make([]string, len(allowedTransitions))
	for i, t := range allowedTransitions {
		validTargets[i] = string(t.To)
	}
	return fmt.Errorf(
		"invalid transition from '%s' to '%s'; valid transitions: %v",
		currentStatus, newStatus, validTargets,
	)
}

// GetAllowedTransitions returns the list of statuses that an order can
// move to from its current status, optionally filtered by role.
func GetAllowedTransitions(currentStatus models.OrderStatus, role models.Role) []models.OrderStatus {
	transitions, exists := transitionMap[currentStatus]
	if !exists {
		return nil
	}
	var result []models.OrderStatus
	for _, t := range transitions {
		if role == "" {
			result = append(result, t.To)
			continue
		}
		for _, allowedRole := range t.AllowedRoles {
			if allowedRole == role {
				result = append(result, t.To)
				break
			}
		}
	}
	return result
}
