package models

// Role represents a user's role in the system.
type Role string

const (
	RoleCustomer   Role = "customer"
	RoleRestaurant Role = "restaurant"
	RoleDriver     Role = "driver"
)

// IsValid checks whether a role string is one of the allowed roles.
func (r Role) IsValid() bool {
	switch r {
	case RoleCustomer, RoleRestaurant, RoleDriver:
		return true
	}
	return false
}

// User represents a registered user (customer, restaurant, or driver).
type User struct {
	ID   string `json:"id" bson:"_id,omitempty"`
	Name string `json:"name" bson:"name"`
	Role Role   `json:"role" bson:"role"`
}

// CreateUserRequest is the payload for registering a new user.
type CreateUserRequest struct {
	Name string `json:"name"`
	Role Role   `json:"role"`
}
