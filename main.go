package main

import (
	"food-delivery-api/db"
	"food-delivery-api/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// MongoDB connection URI — defaults to localhost.
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Connect to MongoDB.
	store, err := db.NewStore(mongoURI)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}
	defer store.Disconnect()

	// Initialize handlers.
	orderHandler := handlers.NewOrderHandler(store)
	userHandler := handlers.NewUserHandler(store)
	menuHandler := handlers.NewMenuHandler(store)

	// Set up router.
	r := mux.NewRouter()

	// --- Public routes (no auth required) ---
	r.HandleFunc("/api/users", userHandler.RegisterUser).Methods("POST")
	r.HandleFunc("/api/users", userHandler.ListUsers).Methods("GET")
	r.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/api/restaurants/{id}/menu", menuHandler.GetMenu).Methods("GET")

	// Health check.
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	// --- Protected routes (auth middleware applied per-handler) ---
	auth := handlers.AuthMiddleware
	r.Handle("/api/orders", auth(http.HandlerFunc(orderHandler.CreateOrder))).Methods("POST")
	r.Handle("/api/orders", auth(http.HandlerFunc(orderHandler.ListOrders))).Methods("GET")
	r.Handle("/api/orders/{id}", auth(http.HandlerFunc(orderHandler.GetOrder))).Methods("GET")
	r.Handle("/api/orders/{id}/status", auth(http.HandlerFunc(orderHandler.UpdateOrderStatus))).Methods("PATCH")
	r.Handle("/api/orders/{id}/history", auth(http.HandlerFunc(orderHandler.GetOrderHistory))).Methods("GET")
	r.Handle("/api/orders/{id}/transitions", auth(http.HandlerFunc(orderHandler.GetAllowedTransitions))).Methods("GET")

	// Menu management (auth required — only restaurant owner).
	r.Handle("/api/restaurants/{id}/menu", auth(http.HandlerFunc(menuHandler.AddMenuItem))).Methods("POST")
	r.Handle("/api/restaurants/{id}/menu/{itemId}", auth(http.HandlerFunc(menuHandler.DeleteMenuItem))).Methods("DELETE")

	// --- Serve frontend static files ---
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	// Start server.
	addr := ":8080"
	log.Printf("🚀 Food Delivery API running on http://localhost%s", addr)
	log.Printf("🌐 Open http://localhost%s in your browser for the dashboard", addr)
	log.Printf("📖 API Endpoints:")
	log.Printf("   POST   /api/users                          - Register user")
	log.Printf("   GET    /api/users                          - List users")
	log.Printf("   GET    /api/users/{id}                     - Get user")
	log.Printf("   GET    /api/restaurants/{id}/menu           - View restaurant menu")
	log.Printf("   POST   /api/restaurants/{id}/menu           - Add menu item (restaurant)")
	log.Printf("   DELETE /api/restaurants/{id}/menu/{itemId}  - Delete menu item")
	log.Printf("   POST   /api/orders                         - Create order (customer)")
	log.Printf("   GET    /api/orders                          - List orders")
	log.Printf("   GET    /api/orders/{id}                     - Get order")
	log.Printf("   PATCH  /api/orders/{id}/status              - Update status")
	log.Printf("   GET    /api/orders/{id}/history             - Status history")
	log.Printf("   GET    /api/orders/{id}/transitions         - Allowed transitions")
	log.Printf("   GET    /health                              - Health check")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
