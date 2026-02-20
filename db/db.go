package db

import (
	"context"
	"fmt"
	"food-delivery-api/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store wraps a MongoDB client and provides CRUD operations.
type Store struct {
	client    *mongo.Client
	db        *mongo.Database
	users     *mongo.Collection
	orders    *mongo.Collection
	menuItems *mongo.Collection
}

// NewStore connects to MongoDB and returns a Store.
func NewStore(mongoURI string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection.
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database("fooddash")
	log.Println("✅ Connected to MongoDB")

	return &Store{
		client:    client,
		db:        db,
		users:     db.Collection("users"),
		orders:    db.Collection("orders"),
		menuItems: db.Collection("menu_items"),
	}, nil
}

// Disconnect closes the MongoDB connection.
func (s *Store) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.client.Disconnect(ctx)
}

// ==================== USER OPERATIONS ====================

// SaveUser inserts or replaces a user document.
func (s *Store) SaveUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.Replace().SetUpsert(true)
	_, err := s.users.ReplaceOne(ctx, bson.M{"_id": user.ID}, user, opts)
	return err
}

// GetUser retrieves a user by ID.
func (s *Store) GetUser(id string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var user models.User
	err := s.users.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return &user, err
}

// ListUsers returns all users, optionally filtered by role.
func (s *Store) ListUsers(roleFilter models.Role) ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{}
	if roleFilter != "" {
		filter["role"] = roleFilter
	}
	cursor, err := s.users.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	if users == nil {
		users = []*models.User{}
	}
	return users, nil
}

// ==================== ORDER OPERATIONS ====================

// SaveOrder inserts or replaces an order document.
func (s *Store) SaveOrder(order *models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.Replace().SetUpsert(true)
	_, err := s.orders.ReplaceOne(ctx, bson.M{"_id": order.ID}, order, opts)
	return err
}

// GetOrder retrieves an order by ID.
func (s *Store) GetOrder(id string) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var order models.Order
	err := s.orders.FindOne(ctx, bson.M{"_id": id}).Decode(&order)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("order not found: %s", id)
	}
	return &order, err
}

// ListOrders returns all orders, optionally filtered by status.
func (s *Store) ListOrders(statusFilter models.OrderStatus) ([]*models.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{}
	if statusFilter != "" {
		filter["status"] = statusFilter
	}
	cursor, err := s.orders.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var orders []*models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	if orders == nil {
		orders = []*models.Order{}
	}
	return orders, nil
}

// ==================== MENU OPERATIONS ====================

// SaveMenuItem inserts or replaces a menu item document.
func (s *Store) SaveMenuItem(item *models.MenuItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.Replace().SetUpsert(true)
	_, err := s.menuItems.ReplaceOne(ctx, bson.M{"_id": item.ID}, item, opts)
	return err
}

// GetMenuItem retrieves a menu item by ID.
func (s *Store) GetMenuItem(id string) (*models.MenuItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var item models.MenuItem
	err := s.menuItems.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("menu item not found: %s", id)
	}
	return &item, err
}

// ListMenuItems returns all menu items for a restaurant.
func (s *Store) ListMenuItems(restaurantID string) ([]*models.MenuItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"restaurant_id": restaurantID}
	cursor, err := s.menuItems.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []*models.MenuItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []*models.MenuItem{}
	}
	return items, nil
}

// DeleteMenuItem removes a menu item by ID.
func (s *Store) DeleteMenuItem(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := s.menuItems.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
