package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func post(url string, body map[string]interface{}, headers map[string]string) map[string]interface{} {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Printf("[%d] %s\n", resp.StatusCode, string(data))
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

func patch(url string, body map[string]interface{}, headers map[string]string) (int, map[string]interface{}) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Printf("[%d] %s\n", resp.StatusCode, string(data))
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return resp.StatusCode, result
}

func get(url string, headers map[string]string) map[string]interface{} {
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Printf("[%d] %s\n", resp.StatusCode, string(data))
	var result map[string]interface{}
	json.Unmarshal(data, &result)
	return result
}

func main() {
	base := "http://localhost:8080"
	passed := 0
	failed := 0

	check := func(name string, condition bool) {
		if condition {
			fmt.Printf("✅ PASS: %s\n", name)
			passed++
		} else {
			fmt.Printf("❌ FAIL: %s\n", name)
			failed++
		}
	}

	// 1. Register users
	fmt.Println("\n=== REGISTER USERS ===")
	customer := post(base+"/api/users", map[string]interface{}{"name": "Alice", "role": "customer"}, nil)
	customerID := customer["id"].(string)
	check("Customer registered", customerID != "")

	restaurant := post(base+"/api/users", map[string]interface{}{"name": "Pizza Palace", "role": "restaurant"}, nil)
	restaurantID := restaurant["id"].(string)
	check("Restaurant registered", restaurantID != "")

	driver := post(base+"/api/users", map[string]interface{}{"name": "Bob Driver", "role": "driver"}, nil)
	driverID := driver["id"].(string)
	check("Driver registered", driverID != "")

	// 2. Create order
	fmt.Println("\n=== CREATE ORDER ===")
	custHeaders := map[string]string{"X-User-ID": customerID, "X-User-Role": "customer"}
	restHeaders := map[string]string{"X-User-ID": restaurantID, "X-User-Role": "restaurant"}
	drvHeaders := map[string]string{"X-User-ID": driverID, "X-User-Role": "driver"}

	order := post(base+"/api/orders", map[string]interface{}{
		"restaurant_id":    restaurantID,
		"items":            []map[string]interface{}{{"name": "Margherita Pizza", "quantity": 2, "price": 12.99}},
		"delivery_address": "123 Main St",
	}, custHeaders)
	orderID := order["id"].(string)
	check("Order created with status PLACED", order["status"] == "PLACED")

	// 3. Test invalid transition: customer trying to confirm
	fmt.Println("\n=== INVALID: CUSTOMER CONFIRMS ===")
	code, _ := patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "CONFIRMED"}, custHeaders)
	check("Customer cannot confirm (403)", code == 403)

	// 4. Test invalid state jump: restaurant skips to DELIVERED
	fmt.Println("\n=== INVALID: SKIP TO DELIVERED ===")
	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "DELIVERED"}, restHeaders)
	check("Cannot skip to DELIVERED (400)", code == 400)

	// 5. Happy path: full lifecycle
	fmt.Println("\n=== HAPPY PATH ===")
	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "CONFIRMED"}, restHeaders)
	check("PLACED → CONFIRMED (200)", code == 200)

	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "PREPARING"}, restHeaders)
	check("CONFIRMED → PREPARING (200)", code == 200)

	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "READY_FOR_PICKUP"}, restHeaders)
	check("PREPARING → READY_FOR_PICKUP (200)", code == 200)

	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "PICKED_UP"}, drvHeaders)
	check("READY_FOR_PICKUP → PICKED_UP (200)", code == 200)

	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "OUT_FOR_DELIVERY"}, drvHeaders)
	check("PICKED_UP → OUT_FOR_DELIVERY (200)", code == 200)

	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "DELIVERED"}, drvHeaders)
	check("OUT_FOR_DELIVERY → DELIVERED (200)", code == 200)

	// 6. Test terminal state: cannot transition from DELIVERED
	fmt.Println("\n=== INVALID: TRANSITION FROM DELIVERED ===")
	code, _ = patch(base+"/api/orders/"+orderID+"/status", map[string]interface{}{"status": "PLACED"}, restHeaders)
	check("Cannot transition from DELIVERED (400)", code == 400)

	// 7. Test cancellation flow
	fmt.Println("\n=== CANCELLATION FLOW ===")
	order2 := post(base+"/api/orders", map[string]interface{}{
		"restaurant_id":    restaurantID,
		"items":            []map[string]interface{}{{"name": "Burger", "quantity": 1, "price": 9.99}},
		"delivery_address": "456 Oak Ave",
	}, custHeaders)
	order2ID := order2["id"].(string)
	code, _ = patch(base+"/api/orders/"+order2ID+"/status", map[string]interface{}{"status": "CANCELLED"}, custHeaders)
	check("Customer cancels PLACED order (200)", code == 200)

	// 8. Check history
	fmt.Println("\n=== ORDER HISTORY ===")
	get(base+"/api/orders/"+orderID+"/history", custHeaders)

	// Summary
	fmt.Printf("\n=== RESULTS: %d passed, %d failed ===\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
