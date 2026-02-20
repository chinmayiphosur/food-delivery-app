# AI Prompts Used

This file documents all AI prompts used during the development of this project, as required by the submission guidelines.

---

## Prompt 1 — Initial Project Generation

```
i want everything to be in go language.
Build a REST API for managing food delivery orders (like Uber Eats or DoorDash). Customers place
orders, restaurants accept and prepare them, and drivers deliver them. The system must handle order
lifecycle state transitions, real-time status updates, and coordinate between three user types. The
challenge is implementing a robust state machine and ensuring valid state transitions.

Deliverables:
- Complete REST API
- State machine diagram showing all transitions
- Database schema with order state tracking
- README with API documentation and state machine explanation
- Document explaining your approach to preventing invalid state transitions
```

## AI Tool Used

- **Antigravity** (Google DeepMind) — Used for code generation, architecture design, and documentation writing.

## What Was AI-Generated vs Human-Directed

| Component | AI Contribution |
|-----------|----------------|
| Architecture design | AI proposed layered architecture; human approved |
| State machine transitions | AI designed 8-state model with role gating; human approved |
| Go source code | AI generated all source files based on approved plan |
| Documentation | AI generated README, design doc, state machine docs |
| Error handling strategy | AI proposed 400 vs 403 distinction for transition errors |
