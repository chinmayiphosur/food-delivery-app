# Design Document — Food Delivery Order Management API

## Overview

This document explains the architectural decisions, state machine design, and approach to preventing invalid state transitions in the Food Delivery Order Management API.

---

## Architecture

The API follows a **layered architecture** with clear separation of concerns:

```
HTTP Request
    │
    ▼
┌─────────────────┐
│   Middleware     │  ← Extracts X-User-ID, X-User-Role headers
├─────────────────┤
│   Handlers      │  ← Input validation, HTTP response codes
├─────────────────┤
│  State Machine  │  ← Business rule enforcement (transitions + roles)
├─────────────────┤
│   Database      │  ← Thread-safe in-memory storage
├─────────────────┤
│   Models        │  ← Data structures shared across layers
└─────────────────┘
```

### Why This Architecture?

1. **State machine as a separate package** — The transition logic is decoupled from HTTP concerns. It can be unit-tested independently and reused if we add gRPC or message queue consumers.
2. **Handlers don't embed business rules** — They delegate all state validation to the state machine, keeping HTTP logic thin.
3. **In-memory store with mutex** — Simplifies deployment (no external DB needed) while remaining thread-safe for concurrent requests.

---

## State Machine Design

### The Problem

Food delivery orders move through a fixed sequence of states. Multiple actors (customers, restaurants, drivers) interact with the same order at different stages. Invalid state changes — like a customer marking an order as delivered, or jumping from PLACED to DELIVERED — must be impossible.

### The Solution: Role-Gated Transition Map

The state machine is implemented as a **static transition map** in `statemachine/statemachine.go`:

```go
var transitionMap = map[OrderStatus][]transition{
    StatusPlaced: {
        {To: StatusConfirmed, AllowedRoles: []Role{RoleRestaurant}},
        {To: StatusCancelled, AllowedRoles: []Role{RoleCustomer}},
    },
    // ... each state explicitly lists its valid transitions
}
```

This design provides **two layers of protection**:

#### Layer 1: Transition Legality
Every valid `(from → to)` pair is explicitly enumerated. If a transition isn't in the map, it's invalid — period. There is no „default allow" behavior.

For example, `PLACED → DELIVERED` is not in the map, so it will be rejected with:
```
400: "invalid transition from 'PLACED' to 'DELIVERED'; valid transitions: [CONFIRMED CANCELLED]"
```

#### Layer 2: Role Authorization
Even if a transition is valid, the caller must have the correct role. Each transition entry lists which roles may perform it.

For example, `PLACED → CONFIRMED` is valid, but only for the `restaurant` role. A customer attempting this gets:
```
403: "role 'customer' is not authorized to transition order from 'PLACED' to 'CONFIRMED'"
```

### Terminal States

`DELIVERED` and `CANCELLED` are **not present as keys** in the transition map. This means:
- No outgoing transitions exist
- `ValidateTransition()` returns: `"no transitions allowed from status 'DELIVERED' (terminal state)"`
- It's structurally impossible to move out of a terminal state

### Complete State Diagram

```
                    ┌──────────┐
                    │  PLACED  │
                    └────┬─────┘
                    ┌────┴─────┐
            ┌───────▼──┐    ┌──▼────────┐
            │CONFIRMED │    │ CANCELLED │
            └────┬─────┘    └───────────┘
                 │   └──────► CANCELLED
            ┌────▼─────┐
            │PREPARING │
            └────┬─────┘
     ┌───────────▼──────────┐
     │  READY_FOR_PICKUP    │
     └───────────┬──────────┘
            ┌────▼─────┐
            │PICKED_UP │
            └────┬─────┘
     ┌───────────▼──────────┐
     │  OUT_FOR_DELIVERY    │
     └───────────┬──────────┘
            ┌────▼─────┐
            │DELIVERED │
            └──────────┘
```

---

## Preventing Invalid State Transitions — Summary

| Mechanism | What It Prevents | HTTP Status |
|-----------|-----------------|-------------|
| **Transition map** | Invalid state jumps (e.g., PLACED → DELIVERED) | 400 Bad Request |
| **Role gating** | Wrong actor performing a transition (e.g., customer confirming) | 403 Forbidden |
| **Terminal states** | Any change after DELIVERED or CANCELLED | 400 Bad Request |
| **Mutex locking** | Race conditions during concurrent updates | N/A (internal) |
| **Type safety** | OrderStatus is a named string type, not a raw string | Compile-time |

### Error Response Strategy

The API distinguishes between two types of failures on the `PATCH /api/orders/{id}/status` endpoint:

- **400 Bad Request** — The transition itself is invalid regardless of who is asking. Example: PLACED → DELIVERED.
- **403 Forbidden** — The transition is valid, but the caller's role doesn't have permission. Example: A customer trying to confirm (which only restaurants may do).

This is achieved by attempting the transition validation with all three roles when the primary validation fails. If any role would succeed, it's a 403; otherwise it's a 400.

---

## Concurrency Safety

The in-memory store uses `sync.RWMutex`:
- **Read operations** (`GetOrder`, `ListOrders`) acquire a read lock — multiple reads can happen concurrently.
- **Write operations** (`SaveOrder`) acquire a write lock — exclusive access, no reads or writes can interleave.

This prevents race conditions like two concurrent status updates both reading the same current state and both succeeding, which could skip a state.

---

## Extensibility

The architecture supports future enhancements:

| Enhancement | What to Change |
|------------|----------------|
| Add a new state (e.g., `REJECTED`) | Add entries to `transitionMap` |
| Add a new role (e.g., `admin`) | Add to `Role` constants, update `transitionMap` |
| Switch to PostgreSQL | Replace `db/db.go` implementation; models/handlers unchanged |
| Add WebSocket notifications | Subscribe to status changes in `UpdateOrderStatus` handler |
| Add order assignment logic | Add driver matching in `READY_FOR_PICKUP → PICKED_UP` transition |

---

## Database Schema (In-Memory)

Although this implementation uses in-memory storage, here is the equivalent relational schema for reference:

```sql
CREATE TABLE users (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    role        VARCHAR(20) NOT NULL CHECK (role IN ('customer', 'restaurant', 'driver'))
);

CREATE TABLE orders (
    id               UUID PRIMARY KEY,
    customer_id      UUID NOT NULL REFERENCES users(id),
    restaurant_id    UUID NOT NULL REFERENCES users(id),
    driver_id        UUID REFERENCES users(id),
    total_amount     DECIMAL(10,2) NOT NULL,
    status           VARCHAR(30) NOT NULL DEFAULT 'PLACED',
    delivery_address TEXT NOT NULL,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id        UUID PRIMARY KEY,
    order_id  UUID NOT NULL REFERENCES orders(id),
    name      VARCHAR(255) NOT NULL,
    quantity  INT NOT NULL,
    price     DECIMAL(10,2) NOT NULL
);

CREATE TABLE status_history (
    id          UUID PRIMARY KEY,
    order_id    UUID NOT NULL REFERENCES orders(id),
    from_status VARCHAR(30),
    to_status   VARCHAR(30) NOT NULL,
    changed_by  UUID NOT NULL REFERENCES users(id),
    role        VARCHAR(20) NOT NULL,
    timestamp   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_status_history_order ON status_history(order_id);
```
